package common

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/shopspring/decimal"
)

// TaskFillTotals 任务维度成交汇总四元组（数量、成交额、手续费、净回款）。
type TaskFillTotals struct {
	Qty    decimal.Decimal // 成交/占用数量
	Amount decimal.Decimal // 成交额
	Fee    decimal.Decimal // 手续费
	Net    decimal.Decimal // 净回款 Amount − Fee
}

// TotalsFromTrades 仅 liquidation_traders 已落库成交（对外展示、终态判定用）。
func TotalsFromTrades(ctx context.Context, fillRepo repository.LiquidationTraderRepo, liquidationID string) (TaskFillTotals, error) {
	var out TaskFillTotals
	if fillRepo == nil {
		return out, nil
	}
	qty, amount, fee, err := fillRepo.AggregateByLiquidationID(ctx, liquidationID)
	if err != nil {
		return out, err
	}
	out.Qty = qty
	out.Amount = amount
	out.Fee = fee
	out.Net = amount.Sub(fee)
	return out, nil
}

// TotalsIncludingInFlight 仅读 liquidation_orders：已成交部分 + 在途整单占用（与 Sync 写单同源，避免 traders 滞后）。
func TotalsIncludingInFlight(
	ctx context.Context,
	orderRepo repository.LiquidationOrderRepo,
	liquidationID string,
	feeRate decimal.Decimal,
) (TaskFillTotals, error) {
	var out TaskFillTotals
	if orderRepo == nil {
		return out, nil
	}
	orders, err := orderRepo.ListByLiquidationID(ctx, liquidationID)
	if err != nil {
		return out, err
	}
	for i := range orders {
		o := &orders[i]
		if isOrderStatusInFlight(o.Status) {
			// 在途：整单 quantity 占抵押；金额 = 已成交 + 未成交按限价估
			out.Qty = out.Qty.Add(o.Quantity)
			out.Amount = out.Amount.Add(orderNotionalIncludingOpen(o))
			continue
		}
		// 终态（filled / cancelled / rejected）：只计实际 filled_qty；0 成交撤单/拒单跳过
		if !o.FilledQty.IsPositive() {
			continue
		}
		out.Qty = out.Qty.Add(o.FilledQty)
		out.Amount = out.Amount.Add(orderFilledNotional(o))
	}
	if feeRate.IsPositive() && out.Amount.IsPositive() {
		out.Fee = out.Amount.Mul(feeRate)
	}
	out.Net = out.Amount.Sub(out.Fee)
	return out, nil
}

func isOrderStatusInFlight(status string) bool {
	switch status {
	case entity.ExchangeOrderStatusNew, entity.ExchangeOrderStatusUnknown, entity.ExchangeOrderStatusPartiallyFilled:
		return true
	default:
		return false
	}
}

func orderFilledNotional(o *entity.LiquidationOrder) decimal.Decimal {
	if !o.FilledQty.IsPositive() {
		return decimal.Zero
	}
	return o.FilledQty.Mul(orderExecPrice(o))
}

// orderNotionalIncludingOpen 在途单：已成交按均价/限价，未成交按限价估名义。
func orderNotionalIncludingOpen(o *entity.LiquidationOrder) decimal.Decimal {
	sum := orderFilledNotional(o)
	openQty := o.Quantity.Sub(o.FilledQty)
	if openQty.IsNegative() {
		openQty = decimal.Zero
	}
	if openQty.IsPositive() {
		sum = sum.Add(o.Price.Mul(openQty))
	}
	return sum
}

func orderExecPrice(o *entity.LiquidationOrder) decimal.Decimal {
	if o.AvgPrice != nil && o.AvgPrice.IsPositive() {
		return *o.AvgPrice
	}
	return o.Price
}

// RemainingDebtAmount trigger_collateral_due_amount − totals.Net（不足 0 则 0）。
func RemainingDebtAmount(task entity.LiquidationTask, totals TaskFillTotals) decimal.Decimal {
	d := task.TriggerCollateralDueAmount.Sub(totals.Net)
	return decimalMaxZero(d)
}

// RemainingCollateralQty trigger_collateral_qty − totals.Qty（不足 0 则 0）。
func RemainingCollateralQty(task entity.LiquidationTask, totals TaskFillTotals) decimal.Decimal {
	d := task.TriggerCollateralQty.Sub(totals.Qty)
	return decimalMaxZero(d)
}

func decimalMaxZero(d decimal.Decimal) decimal.Decimal {
	if d.IsNegative() {
		return decimal.Zero
	}
	return d
}

// SumLiquidatingNotionalBySymbol 在途强平剩余抵押名义（USD 计价用 mid/触发价）。
func SumLiquidatingNotionalBySymbol(
	ctx context.Context,
	taskRepo repository.LiquidationTaskRepo,
	orderRepo repository.LiquidationOrderRepo,
	symbol string,
) (decimal.Decimal, error) {
	if taskRepo == nil {
		return decimal.Zero, nil
	}
	tasks, err := taskRepo.ListRunnable(ctx, 0)
	if err != nil {
		return decimal.Zero, err
	}
	var sum decimal.Decimal
	for _, task := range tasks {
		if ResolveSymbol(task) != symbol {
			continue
		}
		feeRate := task.FeeRate
		totals, err := TotalsIncludingInFlight(ctx, orderRepo, task.LiquidationID, feeRate)
		if err != nil {
			return decimal.Zero, err
		}
		rem := RemainingCollateralQty(task, totals)
		if !rem.IsPositive() {
			continue
		}
		px := task.TriggerMarkPrice
		if task.LastMidPrice != nil && task.LastMidPrice.IsPositive() {
			px = *task.LastMidPrice
		}
		if !px.IsPositive() {
			continue
		}
		sum = sum.Add(rem.Mul(px))
	}
	return sum, nil
}
