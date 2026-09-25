package taskexec

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	httpentity "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/shopspring/decimal"

	"go.uber.org/zap"
)

// execQuoteState 步骤 1 产出：本轮回发单前的行情、剩余量与 symbol 精度。
type execQuoteState struct {
	Mid                    decimal.Decimal // 当前买一价（作 mid / 保护价基准）
	FeeRate                decimal.Decimal // 卖单手续费率（估净回款、反推卖量）
	RemainingDebtAmount    decimal.Decimal // 剩余负债（基于 liquidation_orders 汇总 Net）
	RemainingCollateralQty decimal.Decimal // 剩余可卖抵押（基于 liquidation_orders 汇总 Qty，含在途整单）
	MinQty                 decimal.Decimal // 交易所最小下单量
	Step                   decimal.Decimal // 数量步长（向下取整卖量）
	Tick                   decimal.Decimal // 价格 tick（限价向上取整）
}

// execOrderPlan 步骤 2：保护价、拟卖数量与限价。
type execOrderPlan struct {
	Floor      decimal.Decimal
	NeedSell   decimal.Decimal
	LimitPrice decimal.Decimal
	Depth      *httpentity.BinanceDepth
}

// step0TerminalCheck 无在途单后按 liquidation_traders 汇总判 settled/bad_debt；已终态则结束本轮回。
func (j *TaskExec) step0TerminalCheck(
	ctx context.Context,
	auth httpentity.BinanceAuth,
	task entity.LiquidationTask,
	symbol string,
) bool {
	minQty, _, _ := common.SymbolLotSize(ctx, j.binance, &auth, symbol)
	return j.step6TerminalAfterFill(ctx, task, minQty)
}

// step1QuoteAndRemaining 拉买一价，按 liquidation_orders 算剩余负债/抵押与精度。
func (j *TaskExec) step1QuoteAndRemaining(
	ctx context.Context,
	auth httpentity.BinanceAuth,
	task entity.LiquidationTask,
	symbol string,
) (*execQuoteState, bool) {
	ticker, err := j.binance.GetBookTicker(ctx, &auth, symbol)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-executor-BookTicker", zap.Error(err))
		return nil, false
	}
	bid, _ := decimal.NewFromString(ticker.BidPrice)
	if !bid.IsPositive() {
		return nil, false
	}

	feeRate := decimal.NewFromFloat(conf.GetBinanceFeeRate())
	if task.FeeRate.IsPositive() {
		feeRate = task.FeeRate
	}
	totals, err := common.TotalsIncludingInFlight(ctx, j.orderRepo, task.LiquidationID, feeRate)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-executor-TotalsIncludingInFlight", zap.Error(err))
		return nil, false
	}
	remainingDebtAmount := common.RemainingDebtAmount(task, totals)
	remainingCollateralQty := common.RemainingCollateralQty(task, totals)
	minQty, step, tick := common.SymbolLotSize(ctx, j.binance, &auth, symbol)

	return &execQuoteState{
		Mid:                    bid,
		FeeRate:                feeRate,
		RemainingDebtAmount:    remainingDebtAmount,
		RemainingCollateralQty: remainingCollateralQty,
		MinQty:                 minQty,
		Step:                   step,
		Tick:                   tick,
	}, true
}

// step2PlanSellQty 每轮按档位偏移、剩余负债/抵押与深度上限计算本笔卖量与限价（含第一轮）。
func (j *TaskExec) step2PlanSellQty(
	ctx context.Context,
	auth httpentity.BinanceAuth,
	task entity.LiquidationTask,
	strategy *entity.LiquidationStrategyConfig,
	tierCfg *entity.LiquidationTierConfig,
	symbol string,
	quote *execQuoteState,
) (*execOrderPlan, bool) {
	// 盘口保护价 = 买一价×(1−当前档位盘口偏移)
	offset := decimal.RequireFromString("0.003")
	if tierCfg != nil {
		// 根据档位P，获取买一价偏移比例
		offset = common.BidOffsetForTier(*tierCfg, task.TriggerCollateralAsset)
	}
	floor := quote.Mid.Mul(decimal.NewFromInt(1).Sub(offset))

	// 计算最大下单数量（交易所深度超过floor以上数量的20%）
	depth, _ := j.binance.GetDepth(ctx, &auth, symbol, 100)
	depthSum := common.SumBidDepthAbove(depth, floor)
	orderDepthRatio := decimal.RequireFromString("0.2")
	if strategy != nil && strategy.OrderDepthRatio.IsPositive() {
		orderDepthRatio = strategy.OrderDepthRatio
	}
	depthCap := depthSum.Mul(orderDepthRatio)

	// 计算本次数量
	// x = D ÷ (P × (1 − f))
	effectivePrice := floor.Mul(decimal.NewFromInt(1).Sub(quote.FeeRate))
	if !effectivePrice.IsPositive() {
		return nil, false
	}
	needSell := quote.RemainingDebtAmount.Div(effectivePrice)
	if needSell.GreaterThan(quote.RemainingCollateralQty) {
		needSell = quote.RemainingCollateralQty
	}
	// 本次下单数量和最大下单数量进行比较
	if depthCap.IsPositive() && needSell.GreaterThan(depthCap) {
		needSell = depthCap
	}

	// 数量精度调整
	needSell = common.RoundQtyDown(needSell, quote.Step)

	// 价格精度调整
	limitPrice := common.RoundPriceUp(floor, quote.Tick)

	if !needSell.IsPositive() || needSell.LessThan(quote.MinQty) {
		// 本轮回无法凑够最小卖量：不在这里判终态（仅 trades 汇总在 step0/step6 判 settled/bad_debt）
		return nil, false
	}

	return &execOrderPlan{
		Floor:      floor,
		NeedSell:   needSell,
		LimitPrice: limitPrice,
		Depth:      depth,
	}, true
}

// step3GlobalDepthCap 同 symbol 在途卖量 + 本笔不超过全买盘比例；超限则告警并跳过本轮回。
func (j *TaskExec) step3GlobalDepthCap(
	ctx context.Context,
	strategy *entity.LiquidationStrategyConfig,
	symbol string,
	task entity.LiquidationTask,
	plan *execOrderPlan,
) bool {
	globalCap := decimal.RequireFromString("0.30")
	if strategy != nil && strategy.GlobalDepthCapRatio.IsPositive() {
		globalCap = strategy.GlobalDepthCapRatio
	}
	depthAll := common.SumAllBidDepth(plan.Depth)
	inFlightSymbolQty, _ := j.orderRepo.SumInFlightSellQtyBySymbol(ctx, symbol)
	if depthAll.IsPositive() && inFlightSymbolQty.Add(plan.NeedSell).GreaterThan(depthAll.Mul(globalCap)) {
		common.AlertLiquidation(ctx, "global_depth_cap", "在途卖出量超过全局买盘深度上限，暂停发单", map[string]string{
			"symbol": symbol, "liquidation_id": task.LiquidationID,
			"in_flight": inFlightSymbolQty.String(), "planned": plan.NeedSell.String(),
		})
		return false
	}
	return true
}

// step4PersistBatchAndOrder 创建子单记录（尚未请求交易所）。
func (j *TaskExec) step4PersistBatchAndOrder(
	ctx context.Context,
	task entity.LiquidationTask,
	tier, symbol string,
	quote *execQuoteState,
	plan *execOrderPlan,
) (*entity.LiquidationOrder, bool) {
	batchSeq, err := j.orderRepo.NextBatchSeq(ctx, task.LiquidationID)
	if err != nil {
		return nil, false
	}
	batchNo := fmt.Sprintf("%s-B%d", task.LiquidationID, batchSeq)
	tierCopy := tier

	clientOrderID := common.NewClientOrderID(task.LiquidationID, batchSeq)
	submitStr := fmt.Sprintf(`{"symbol":"%s","qty":"%s","price":"%s"}`, symbol, common.FormatDecimal(plan.NeedSell), common.FormatDecimal(plan.LimitPrice))
	snapshotStr, err := buildExecSnapshotJSON(task, tier, symbol, batchNo, batchSeq, quote, plan)
	if err != nil {
		return nil, false
	}
	floorCopy := plan.Floor
	now := time.Now()
	row := &entity.LiquidationOrder{
		LiquidationID: task.LiquidationID,
		BatchNo:       batchNo,
		ClientOrderID: clientOrderID,
		ExchangeCode:  entity.ExchangeCodeBinance,
		Symbol:        symbol,
		Side:          httpentity.BinanceSideSell,
		OrderType:     "LIMIT_IOC",
		Tier:          &tierCopy,
		FloorPrice:    &floorCopy,
		Price:         plan.LimitPrice,
		Quantity:      plan.NeedSell,
		Status:        entity.ExchangeOrderStatusNew,
		SubmitRaw:     &submitStr,
		ExecSnapshot:  &snapshotStr,
		SubmittedAt:   &now,
	}
	if err = j.orderRepo.Create(ctx, j.db, row); err != nil {
		return nil, false
	}
	common.EmitTaskEvent(ctx, j.db, j.eventRepo, task.LiquidationID, entity.TaskEventOrderSubmit, "", "", clientOrderID, nil)
	return row, true
}

// step5SubmitLimitIOC 提交 Binance IOC，失败处置；成功则 Apply 成交。
func (j *TaskExec) step5SubmitLimitIOC(
	ctx context.Context,
	auth httpentity.BinanceAuth,
	task entity.LiquidationTask,
	strategy *entity.LiquidationStrategyConfig,
	row *entity.LiquidationOrder,
	symbol string,
	plan *execOrderPlan,
) bool {
	now := time.Now()
	if row.SubmittedAt != nil {
		now = *row.SubmittedAt
	}

	bo, postErr := j.binance.CreateSpotOrder(ctx, auth, httpentity.BinanceCreateOrderReq{
		Symbol:           symbol,
		Side:             httpentity.BinanceSideSell,
		Type:             httpentity.BinanceOrderTypeLimit,
		TimeInForce:      httpentity.BinanceTIFIOC,
		Quantity:         common.FormatDecimal(plan.NeedSell),
		Price:            common.FormatDecimal(plan.LimitPrice),
		NewClientOrderID: row.ClientOrderID,
	})
	if postErr != nil {
		if common.IsInsufficientBalance(postErr) {
			common.PauseTask(ctx, j.db, j.taskRepo, j.eventRepo, task, "可用余额不足")
			common.AlertLiquidation(ctx, "insufficient_balance", postErr.Error(), map[string]string{"liquidation_id": task.LiquidationID})
			_ = j.orderRepo.Update(ctx, j.db, row.ID, map[string]interface{}{
				"status": entity.ExchangeOrderStatusRejected, "failure_reason": postErr.Error(), "finished_at": now,
			})
			return false
		}
		streak := task.OrderFailStreak + 1
		updates := map[string]interface{}{
			"order_fail_streak": streak,
		}
		if strategy != nil && streak >= strategy.ConsecutiveOrderFailPause {
			common.PauseTask(ctx, j.db, j.taskRepo, j.eventRepo, task, "连续下单失败")
			return false
		}
		st := entity.ExchangeOrderStatusUnknown
		reason := postErr.Error()
		var binanceError *httpentity.BinanceError
		if errors.As(postErr, &binanceError) {
			st = entity.ExchangeOrderStatusRejected
			common.AlertLiquidation(ctx, "order_rejected", reason, map[string]string{"liquidation_id": task.LiquidationID})
		}
		_ = j.orderRepo.Update(ctx, j.db, row.ID, map[string]interface{}{
			"status": st, "failure_reason": reason, "finished_at": now,
		})
		_ = j.taskRepo.UpdateByLiquidationID(ctx, j.db, task.LiquidationID, updates)
		return false
	}

	_ = j.taskRepo.UpdateByLiquidationID(ctx, j.db, task.LiquidationID, map[string]interface{}{"order_fail_streak": 0})
	if err := j.binanceSync.ApplyBinanceOrder(ctx, auth, row, &task, bo); err != nil {
		notification.SendErrorLog(ctx, "liquidation-executor-ApplyOrder", zap.Error(err))
	}
	return true
}

// step6TerminalAfterFill 终态唯一依据：liquidation_traders 汇总（step0 发单前、step6 成交后各调一次）。
func (j *TaskExec) step6TerminalAfterFill(ctx context.Context, task entity.LiquidationTask, minQty decimal.Decimal) bool {
	taskAfter, _ := j.taskRepo.FindByLiquidationID(ctx, task.LiquidationID)
	if taskAfter == nil {
		return false
	}
	totals, err := common.TotalsFromTrades(ctx, j.fillRepo, taskAfter.LiquidationID)
	if err != nil {
		return false
	}
	rd := common.RemainingDebtAmount(*taskAfter, totals)
	rc := common.RemainingCollateralQty(*taskAfter, totals)
	return j.checkTerminal(ctx, *taskAfter, rd, rc, minQty)
}
