package taskexec

import (
	"context"
	"sort"
	"strings"
	"time"

	appcommon "github.com/gph-tech/fgmm-strategy-bitfinex/application/common"
	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/orderreconcile"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/tasktiercalc"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	httpentity "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	httprepo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/shopspring/decimal"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TaskExec 扫描 liquidating：读 DB 档位，按 P 优先级执行 Binance IOC。
type TaskExec struct {
	db          *gorm.DB
	taskRepo    repository.LiquidationTaskRepo
	orderRepo   repository.LiquidationOrderRepo
	fillRepo    repository.LiquidationTraderRepo
	configRepo  repository.LiquidationExecutionConfigRepo
	eventRepo   repository.LiquidationTaskEventRepo
	binance     httprepo.BinanceRepo
	binanceSync     common.BinanceSync
	tierCalc        *tasktiercalc.TaskTierCalc
	orderReconcile  *orderreconcile.OrderReconcile
	platformSvc     *coreservice.PlatformConfigService
}

// NewTaskExec 注入依赖。
func NewTaskExec(
	db *gorm.DB,
	taskRepo repository.LiquidationTaskRepo,
	orderRepo repository.LiquidationOrderRepo,
	fillRepo repository.LiquidationTraderRepo,
	configRepo repository.LiquidationExecutionConfigRepo,
	eventRepo repository.LiquidationTaskEventRepo,
	binance httprepo.BinanceRepo,
	tierCalc *tasktiercalc.TaskTierCalc,
	orderReconcile *orderreconcile.OrderReconcile,
	platformSvc *coreservice.PlatformConfigService,
) *TaskExec {
	return &TaskExec{
		db:             db,
		taskRepo:       taskRepo,
		orderRepo:      orderRepo,
		fillRepo:       fillRepo,
		configRepo:     configRepo,
		eventRepo:      eventRepo,
		binance:        binance,
		binanceSync:    common.NewBinanceSync(db, binance, orderRepo, fillRepo, taskRepo, eventRepo, configRepo),
		tierCalc:       tierCalc,
		orderReconcile: orderReconcile,
		platformSvc:    platformSvc,
	}
}

// sortRunnableByTier P0 优先于 P1、P2；同档按 updated_at 先后。
func sortRunnableByTier(tasks []entity.LiquidationTask) {
	sort.SliceStable(tasks, func(i, j int) bool {
		ri := appcommon.TierEmergencyRank(common.CurrentTier(tasks[i]))
		rj := appcommon.TierEmergencyRank(common.CurrentTier(tasks[j]))
		if ri != rj {
			return ri < rj
		}
		return tasks[i].UpdatedAt.Before(tasks[j].UpdatedAt)
	})
}

// Run 先重算档位与 mid 快照，再对账 unknown/超时 new 子单，最后拉 liquidating 任务并按优先级逐单发 IOC。
func (j *TaskExec) Run(ctx context.Context) {
	// tierCalc 会写 last_mid_price_at；本轮回内视为新鲜，避免多任务顺序 runOne 超过 stale_quote_sec 误报。
	midBatchStartedAt := time.Now()
	j.tierCalc.Run(ctx)
	if j.orderReconcile != nil {
		j.orderReconcile.Run(ctx)
	}

	auth, err := j.platformSvc.BinanceSpotAuth(ctx)
	if err != nil {
		logger.WarnZ(ctx, "liquidation-task_exec-BinanceDisabled", zap.String("reason", err.Error()))
		return
	}

	// 获取所有在清算状态的任务
	tasks, err := j.taskRepo.ListRunnable(ctx, 0)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-task_exec-ListRunnable", zap.Error(err))
		return
	}

	// 基于liquidation_task.current_tier排序，然后顺序执行
	sortRunnableByTier(tasks)
	for _, task := range tasks {
		j.runOne(ctx, auth, task, midBatchStartedAt)
	}
}

// runOne 单任务一轮：前置检查后依次 step1…step6（见 exec_round.go）。
func (j *TaskExec) runOne(ctx context.Context, auth httpentity.BinanceAuth, task entity.LiquidationTask, midBatchStartedAt time.Time) {
	if task.Status == entity.LiquidationStatusPaused {
		return
	}

	// 读取liquidation_config_strategy配置
	strategy, _ := j.configRepo.GetStrategy(ctx)

	// 获取P档位的配置
	tiers := common.LoadTierConfigs(ctx, j.configRepo)
	tier := common.CurrentTier(task)
	tierCfg := common.TierConfigByName(tiers, tier)

	// 还有没有「没结束」的交易所卖单
	inFlight, err := j.orderRepo.HasInFlightByLiquidationID(ctx, task.LiquidationID)
	if err != nil {
		return
	}
	if inFlight {
		// 超过一定时间10S，报警并取消挂单
		j.alertStaleInFlight(ctx, auth, task, strategy)
		return
	}

	symbol := common.ResolveSymbol(task)
	if symbol == "" {
		notification.SendErrorLog(ctx, "liquidation-executor-MissingSymbol", zap.String("liquidation_id", task.LiquidationID))
		return
	}

	// 行情快照过旧则跳过（默认 3s）；本轮回开头 tierCalc 已刷新的 mid 不算 stale。
	if isStaleQuote(task, strategy, midBatchStartedAt) {
		common.AlertLiquidation(ctx, "stale_quote", "任务行情快照超过 stale_quote_sec 未更新，跳过本轮回发单", map[string]string{
			"liquidation_id": task.LiquidationID,
			"symbol":         symbol,
		})
		return
	}

	// 0. 无在途单后先判终态（trades 汇总，与 step6 同一套）
	if j.step0TerminalCheck(ctx, auth, task, symbol) {
		return
	}

	// 1. 行情与剩余负债/抵押（orders 汇总，供本轮回发单）
	quote, ok := j.step1QuoteAndRemaining(ctx, auth, task, symbol)
	if !ok || quote == nil {
		return
	}

	// 2. 计算：卖量与限价
	plan, ok := j.step2PlanSellQty(ctx, auth, task, strategy, tierCfg, symbol, quote)
	if !ok || plan == nil {
		return
	}

	// 3. 卖出数量判断：所有在途清算任务的卖出总量 ≤ 该币对当前买盘深度 × 全局上限比例（默认30%）
	if !j.step3GlobalDepthCap(ctx, strategy, symbol, task, plan) {
		return
	}

	// 4. 创建子单记录
	row, ok := j.step4PersistBatchAndOrder(ctx, task, tier, symbol, quote, plan)
	if !ok || row == nil {
		return
	}

	// 5. 提交 Binance LIMIT IOC，对齐成交
	if !j.step5SubmitLimitIOC(ctx, auth, task, strategy, row, symbol, plan) {
		return
	}

	// 6. 成交对齐后再判 settled / bad_debt
	_ = j.step6TerminalAfterFill(ctx, task, quote.MinQty)
}

// checkTerminal 仅配合 TotalsFromTrades 使用：债务清零 → settled；抵押不足仍有债 → bad_debt。
func (j *TaskExec) checkTerminal(ctx context.Context, task entity.LiquidationTask, remainingDebtAmount, remainingCollateralQty, minQty decimal.Decimal) bool {
	if remainingDebtAmount.IsZero() {
		j.terminalSettled(ctx, task)
		return true
	}
	if remainingCollateralQty.LessThan(minQty) && remainingDebtAmount.IsPositive() {
		j.terminalBadDebt(ctx, task)
		return true
	}
	if remainingDebtAmount.IsPositive() && remainingCollateralQty.IsZero() {
		j.terminalBadDebt(ctx, task)
		return true
	}
	return false
}

// terminalSettled 债务已覆盖，待 callback。
func (j *TaskExec) terminalSettled(ctx context.Context, task entity.LiquidationTask) {
	now := time.Now()
	updates := map[string]interface{}{
		"status": entity.LiquidationStatusSettled, "end_reason": entity.EndReasonDebtCleared, "finished_at": now,
	}
	if strings.TrimSpace(task.CallbackURL) != "" {
		updates["callback_status"] = entity.CallbackStatusPending
	}
	_ = j.taskRepo.UpdateByLiquidationID(ctx, j.db, task.LiquidationID, updates)
	common.EmitTaskEvent(ctx, j.db, j.eventRepo, task.LiquidationID, entity.TaskEventSettled, task.Status, entity.LiquidationStatusSettled, "", nil)
}

// terminalBadDebt 抵押卖尽仍欠债。
func (j *TaskExec) terminalBadDebt(ctx context.Context, task entity.LiquidationTask) {
	now := time.Now()
	updates := map[string]interface{}{
		"status": entity.LiquidationStatusBadDebt, "end_reason": entity.EndReasonCollateralExhausted, "finished_at": now,
	}
	if strings.TrimSpace(task.CallbackURL) != "" {
		updates["callback_status"] = entity.CallbackStatusPending
	}
	_ = j.taskRepo.UpdateByLiquidationID(ctx, j.db, task.LiquidationID, updates)
	common.EmitTaskEvent(ctx, j.db, j.eventRepo, task.LiquidationID, entity.TaskEventBadDebt, task.Status, entity.LiquidationStatusBadDebt, "", nil)
	common.AlertLiquidation(ctx, "bad_debt", "抵押耗尽仍有负债", map[string]string{"liquidation_id": task.LiquidationID})
}

// alertStaleInFlight 挂单超时：order_stale 告警，查单对齐后向交易所撤单并再次同步。
func (j *TaskExec) alertStaleInFlight(ctx context.Context, auth httpentity.BinanceAuth, task entity.LiquidationTask, strategy *entity.LiquidationStrategyConfig) {
	maxSec := uint(300)
	if strategy != nil && strategy.MaxSingleOrderDurationSec > 0 {
		maxSec = strategy.MaxSingleOrderDurationSec
	}
	orders, err := j.orderRepo.ListInFlightByLiquidationID(ctx, task.LiquidationID)
	if err != nil {
		return
	}
	threshold := time.Duration(maxSec) * time.Second
	now := time.Now()
	taskCopy := task
	for i := range orders {
		o := &orders[i]
		if o.SubmittedAt == nil {
			continue
		}
		if now.Sub(*o.SubmittedAt) <= threshold {
			continue
		}
		common.AlertLiquidation(ctx, "order_stale", "单一卖出订单执行超时，发起撤单", map[string]string{
			"liquidation_id":  task.LiquidationID,
			"client_order_id": o.ClientOrderID,
		})
		if err := j.binanceSync.SyncOrder(ctx, auth, o, &taskCopy); err != nil {
			notification.SendErrorLog(ctx, "liquidation-task_exec-staleSync",
				zap.String("liquidation_id", task.LiquidationID),
				zap.String("client_order_id", o.ClientOrderID),
				zap.Error(err))
		}
		staleRow, _ := j.orderRepo.FindByClientOrderID(ctx, o.ClientOrderID)
		if staleRow == nil || !isExchangeOrderInFlight(staleRow.Status) {
			continue
		}
		// 取消订单
		if _, err = j.binance.CancelSpotOrder(ctx, auth, staleRow.Symbol, staleRow.ClientOrderID); err != nil {
			notification.SendErrorLog(ctx, "liquidation-task_exec-staleCancel",
				zap.String("liquidation_id", task.LiquidationID),
				zap.String("client_order_id", o.ClientOrderID),
				zap.Error(err))
			continue
		}
		// 同步订单结果
		if err = j.binanceSync.SyncOrder(ctx, auth, staleRow, &taskCopy); err != nil {
			notification.SendErrorLog(ctx, "liquidation-task_exec-staleSyncAfterCancel",
				zap.String("liquidation_id", task.LiquidationID),
				zap.String("client_order_id", staleRow.ClientOrderID),
				zap.Error(err))
		}
	}
}

// isStaleQuote last_mid_price_at 超过 stale_quote_sec；本轮回 tierCalc 之后写入的快照除外。
func isStaleQuote(task entity.LiquidationTask, strategy *entity.LiquidationStrategyConfig, midBatchStartedAt time.Time) bool {
	if task.LastMidPriceAt == nil || strategy == nil || strategy.StaleQuoteSec == 0 {
		return false
	}
	at := *task.LastMidPriceAt
	if !at.Before(midBatchStartedAt) {
		return false
	}
	return time.Since(at) > time.Duration(strategy.StaleQuoteSec)*time.Second
}

func isExchangeOrderInFlight(status string) bool {
	switch status {
	case entity.ExchangeOrderStatusNew, entity.ExchangeOrderStatusUnknown, entity.ExchangeOrderStatusPartiallyFilled:
		return true
	default:
		return false
	}
}
