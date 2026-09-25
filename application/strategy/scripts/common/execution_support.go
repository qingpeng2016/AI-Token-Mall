package common

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	appcommon "github.com/gph-tech/fgmm-strategy-bitfinex/application/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	httpentity "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	httprepo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/shopspring/decimal"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type symbolFilterCache struct {
	mu sync.RWMutex
	m  map[string]cachedFilter
}

type cachedFilter struct {
	minQty   decimal.Decimal
	stepSize decimal.Decimal
	tickSize decimal.Decimal
	at       time.Time
}

var globalFilterCache = symbolFilterCache{m: make(map[string]cachedFilter)}

// SymbolLotSize 交易对 minQty/step/tick（带 1h 内存缓存）。
func SymbolLotSize(ctx context.Context, binance httprepo.BinanceRepo, auth *httpentity.BinanceAuth, symbol string) (minQty, step, tick decimal.Decimal) {
	return globalFilterCache.get(ctx, binance, auth, symbol)
}

// get 读缓存或拉 exchangeInfo。
func (c *symbolFilterCache) get(ctx context.Context, binance httprepo.BinanceRepo, auth *httpentity.BinanceAuth, symbol string) (minQty, step, tick decimal.Decimal) {
	minQty = decimal.RequireFromString("0.00001000")
	step = decimal.RequireFromString("0.00001000")
	tick = decimal.RequireFromString("0.01000000")
	c.mu.RLock()
	if v, ok := c.m[symbol]; ok && time.Since(v.at) < time.Hour {
		c.mu.RUnlock()
		return v.minQty, v.stepSize, v.tickSize
	}
	c.mu.RUnlock()
	info, err := binance.GetExchangeInfo(ctx, auth, symbol)
	if err != nil || info == nil {
		return minQty, step, tick
	}
	for _, f := range info.Filters {
		switch f.FilterType {
		case "LOT_SIZE":
			if f.MinQty != "" {
				minQty, _ = decimal.NewFromString(f.MinQty)
			}
			if f.StepSize != "" {
				step, _ = decimal.NewFromString(f.StepSize)
			}
		case "PRICE_FILTER":
			if f.TickSize != "" {
				tick, _ = decimal.NewFromString(f.TickSize)
			}
		}
	}
	c.mu.Lock()
	c.m[symbol] = cachedFilter{minQty: minQty, stepSize: step, tickSize: tick, at: time.Now()}
	c.mu.Unlock()
	return minQty, step, tick
}

// RoundQtyDown 按 step 向下取整数量。
func RoundQtyDown(q, step decimal.Decimal) decimal.Decimal {
	if !step.IsPositive() {
		return q.Truncate(8)
	}
	return q.Div(step).Floor().Mul(step)
}

// RoundPriceUp 按 tick 向上取整价格。
func RoundPriceUp(p, tick decimal.Decimal) decimal.Decimal {
	if !tick.IsPositive() {
		return p
	}
	return p.Div(tick).Ceil().Mul(tick)
}

// SumAllBidDepth 买盘总量（全档）。
func SumAllBidDepth(depth *httpentity.BinanceDepth) decimal.Decimal {
	sum := decimal.Zero
	if depth == nil {
		return sum
	}
	for _, row := range depth.Bids {
		if len(row) < 2 {
			continue
		}
		qty, err := decimal.NewFromString(row[1])
		if err != nil {
			continue
		}
		sum = sum.Add(qty)
	}
	return sum
}

// SumBidDepthAbove floor 及以上价位的买盘量。
func SumBidDepthAbove(depth *httpentity.BinanceDepth, floor decimal.Decimal) decimal.Decimal {
	sum := decimal.Zero
	if depth == nil {
		return sum
	}
	for _, row := range depth.Bids {
		if len(row) < 2 {
			continue
		}
		price, err1 := decimal.NewFromString(row[0])
		qty, err2 := decimal.NewFromString(row[1])
		if err1 != nil || err2 != nil {
			continue
		}
		if price.GreaterThanOrEqual(floor) {
			sum = sum.Add(qty)
		}
	}
	return sum
}

// HardBoundaryPct 硬边界比例（按 BTC/ETH）。
func HardBoundaryPct(strategy *entity.LiquidationStrategyConfig, asset string) decimal.Decimal {
	if strategy == nil {
		return decimal.RequireFromString("0.02")
	}
	switch appcommon.NormalizeCollateralAsset(asset) {
	case appcommon.CollateralAssetBTC:
		return strategy.HardBoundaryBTCPct
	case appcommon.CollateralAssetETH:
		return strategy.HardBoundaryETHPct
	default:
		return strategy.HardBoundaryDefaultPct
	}
}

// EmitTaskEvent 写 liquidation_task 事件流水。
func EmitTaskEvent(ctx context.Context, db *gorm.DB, repo repository.LiquidationTaskEventRepo, liquidationID, eventType, from, to, msg string, extra any) {
	if repo == nil {
		return
	}
	var extraStr *string
	if extra != nil {
		b, _ := json.Marshal(extra)
		s := string(b)
		extraStr = &s
	}
	var fromP, toP, msgP *string
	if from != "" {
		fromP = &from
	}
	if to != "" {
		toP = &to
	}
	if msg != "" {
		msgP = &msg
	}
	_ = repo.Create(ctx, db, &entity.LiquidationTaskEvent{
		LiquidationID: liquidationID,
		EventType:     eventType,
		FromStatus:    fromP,
		ToStatus:      toP,
		Message:       msgP,
		Extra:         extraStr,
	})
}

// AlertLiquidation 强平告警（通知 + 日志）。
func AlertLiquidation(ctx context.Context, title, msg string, fields map[string]string) {
	zf := []zap.Field{zap.String("title", title), zap.String("message", msg)}
	for k, v := range fields {
		zf = append(zf, zap.String(k, v))
	}
	notification.SendErrorLog(ctx, "liquidation-alert", zf...)
}

// PauseTask 置 paused 并告警（不主动撤单，由 task_cancel 处理）。
func PauseTask(ctx context.Context, db *gorm.DB, taskRepo repository.LiquidationTaskRepo, eventRepo repository.LiquidationTaskEventRepo, task entity.LiquidationTask, reason string) {
	_ = taskRepo.UpdateByLiquidationID(ctx, db, task.LiquidationID, map[string]interface{}{
		"status": entity.LiquidationStatusPaused,
	})
	EmitTaskEvent(ctx, db, eventRepo, task.LiquidationID, entity.TaskEventPause, task.Status, entity.LiquidationStatusPaused, reason, nil)
	AlertLiquidation(ctx, "liquidation-paused", reason, map[string]string{"liquidation_id": task.LiquidationID})
}
