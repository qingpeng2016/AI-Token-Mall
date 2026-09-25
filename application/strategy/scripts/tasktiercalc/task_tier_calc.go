package tasktiercalc

import (
	"context"
	"time"

	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	httprepo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/shopspring/decimal"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TaskTierCalc 扫描 liquidating 任务，重算 P0/P1/P2 并写入 liquidation_task.current_tier。
type TaskTierCalc struct {
	db         *gorm.DB
	taskRepo   repository.LiquidationTaskRepo
	orderRepo  repository.LiquidationOrderRepo
	configRepo repository.LiquidationExecutionConfigRepo
	eventRepo  repository.LiquidationTaskEventRepo
	binance     httprepo.BinanceRepo
	platformSvc *coreservice.PlatformConfigService
}

// NewTaskTierCalc 注入依赖。
func NewTaskTierCalc(
	db *gorm.DB,
	taskRepo repository.LiquidationTaskRepo,
	orderRepo repository.LiquidationOrderRepo,
	configRepo repository.LiquidationExecutionConfigRepo,
	eventRepo repository.LiquidationTaskEventRepo,
	binance httprepo.BinanceRepo,
	platformSvc *coreservice.PlatformConfigService,
) *TaskTierCalc {
	return &TaskTierCalc{
		db:          db,
		taskRepo:    taskRepo,
		orderRepo:   orderRepo,
		configRepo:  configRepo,
		eventRepo:   eventRepo,
		binance:     binance,
		platformSvc: platformSvc,
	}
}

// Run 拉全部 liquidating 任务，按 symbol 取买一价后更新档位与 mid 快照。
func (j *TaskTierCalc) Run(ctx context.Context) {
	auth, err := j.platformSvc.BinanceSpotAuth(ctx)
	if err != nil {
		logger.WarnZ(ctx, "liquidation-task_tier_calc-BinanceDisabled", zap.String("reason", err.Error()))
		return
	}

	tasks, err := j.taskRepo.ListRunnable(ctx, 0)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-task_tier_calc-ListRunnable", zap.Error(err))
		return
	}
	if len(tasks) == 0 {
		return
	}

	tiers := common.LoadTierConfigs(ctx, j.configRepo)
	now := time.Now()
	midBySymbol := make(map[string]decimal.Decimal)

	for _, task := range tasks {
		symbol := common.ResolveSymbol(task)
		if symbol == "" {
			continue
		}
		mid, ok := midBySymbol[symbol]
		if !ok {
			ticker, err := j.binance.GetBookTicker(ctx, &auth, symbol)
			if err != nil {
				notification.SendErrorLog(ctx, "liquidation-task_tier_calc-BookTicker", zap.Error(err), zap.String("symbol", symbol))
				continue
			}
			bid, _ := decimal.NewFromString(ticker.BidPrice)
			if !bid.IsPositive() {
				continue
			}
			mid = bid
			midBySymbol[symbol] = mid
		}

		iocStreak, err := j.orderRepo.CountConsecutiveIOCUnfilled(ctx, task.LiquidationID)
		if err != nil {
			continue
		}
		tier, explain := common.ResolveTierForTaskWithExplain(task, tiers, mid, iocStreak)
		prev := common.CurrentTier(task)
		reasonJSON, err := common.MarshalTierResolutionExplain(explain)
		if err != nil {
			notification.SendErrorLog(ctx, "liquidation-task_tier_calc-ReasonJSON", zap.Error(err), zap.String("liquidation_id", task.LiquidationID))
			continue
		}
		updates := map[string]interface{}{
			"current_tier":              tier,
			"current_tier_reason_json":  reasonJSON,
			"last_mid_price":            mid,
			"last_mid_price_at":         now,
		}
		if err := j.taskRepo.UpdateByLiquidationID(ctx, j.db, task.LiquidationID, updates); err != nil {
			notification.SendErrorLog(ctx, "liquidation-task_tier_calc-Update", zap.Error(err), zap.String("liquidation_id", task.LiquidationID))
			continue
		}
		if tier != prev {
			common.EmitTaskEvent(ctx, j.db, j.eventRepo, task.LiquidationID, entity.TaskEventTierChanged, prev, tier, "", nil)
		}
	}
}
