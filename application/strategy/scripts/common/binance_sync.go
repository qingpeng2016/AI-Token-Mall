package common

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	httpentity "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	httprepo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/shopspring/decimal"

	"gorm.io/gorm"
)

// BinanceSync 查单、落成交、汇总任务（executor / reconcile / cancel 共用）。
type BinanceSync struct {
	db         *gorm.DB
	binance    httprepo.BinanceRepo
	orderRepo  repository.LiquidationOrderRepo
	fillRepo   repository.LiquidationTraderRepo
	taskRepo   repository.LiquidationTaskRepo
	eventRepo  repository.LiquidationTaskEventRepo
	configRepo repository.LiquidationExecutionConfigRepo
}

// NewBinanceSync 注入依赖。
func NewBinanceSync(
	db *gorm.DB,
	binance httprepo.BinanceRepo,
	orderRepo repository.LiquidationOrderRepo,
	fillRepo repository.LiquidationTraderRepo,
	taskRepo repository.LiquidationTaskRepo,
	eventRepo repository.LiquidationTaskEventRepo,
	configRepo repository.LiquidationExecutionConfigRepo,
) BinanceSync {
	return BinanceSync{
		db: db, binance: binance, orderRepo: orderRepo, fillRepo: fillRepo,
		taskRepo: taskRepo, eventRepo: eventRepo, configRepo: configRepo,
	}
}

// SyncOrder GetSpotOrder 后 ApplyBinanceOrder。
func (s *BinanceSync) SyncOrder(ctx context.Context, auth httpentity.BinanceAuth, row *entity.LiquidationOrder, task *entity.LiquidationTask) error {
	q := httpentity.BinanceGetOrderQuery{Symbol: row.Symbol, OrigClientOrderID: row.ClientOrderID}
	if row.ExchangeOrderID != nil && *row.ExchangeOrderID != "" {
		if oid, err := strconv.ParseInt(*row.ExchangeOrderID, 10, 64); err == nil {
			q.OrderID = oid
		}
	}
	bo, err := s.binance.GetSpotOrder(ctx, auth, q)
	if err != nil {
		return err
	}
	return s.ApplyBinanceOrder(ctx, auth, row, task, bo)
}

// ApplyBinanceOrder 先拉 myTrades，再事务内更新 liquidation_orders 并写入 liquidation_traders。
func (s *BinanceSync) ApplyBinanceOrder(ctx context.Context, auth httpentity.BinanceAuth, row *entity.LiquidationOrder, task *entity.LiquidationTask, bo *httpentity.BinanceOrder) error {
	status := MapBinanceOrderStatus(bo.Status)
	filled, _ := decimal.NewFromString(bo.ExecutedQty)
	now := time.Now()
	raw := string(bo.Raw)
	exID := strconv.FormatInt(bo.OrderID, 10)

	updates := map[string]interface{}{
		"status":            status,
		"filled_qty":        filled,
		"exchange_order_id": exID,
		"last_report_raw":   raw,
	}
	if filled.IsPositive() {
		cumQuote, _ := decimal.NewFromString(bo.CummulativeQuoteQty)
		if cumQuote.IsPositive() {
			avg := cumQuote.Div(filled)
			updates["avg_price"] = avg
		}
	}
	if status == entity.ExchangeOrderStatusFilled || status == entity.ExchangeOrderStatusCancelled || status == entity.ExchangeOrderStatusRejected {
		updates["finished_at"] = now
	}

	var trades []httpentity.BinanceTrade
	if bo.OrderID != 0 && filled.IsPositive() {
		var err error
		trades, err = s.binance.GetMyTrades(ctx, auth, httpentity.BinanceMyTradesQuery{
			Symbol:  row.Symbol,
			OrderID: bo.OrderID,
			Limit:   500,
		})
		if err != nil {
			return err
		}
	}

	row.FilledQty = filled
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.orderRepo.Update(ctx, tx, row.ID, updates); err != nil {
			return err
		}
		if len(trades) == 0 {
			return nil
		}
		return s.persistTrades(ctx, tx, row, trades, task)
	})
}

// persistTrades 事务内写入成交；触及硬边界则同事务内任务 failed。
func (s *BinanceSync) persistTrades(ctx context.Context, tx *gorm.DB, row *entity.LiquidationOrder, trades []httpentity.BinanceTrade, task *entity.LiquidationTask) error {
	strategy, _ := s.configRepo.GetStrategy(ctx)
	var taskSnap entity.LiquidationTask
	if task != nil {
		taskSnap = *task
	} else if s.taskRepo != nil {
		if t, _ := s.taskRepo.FindByLiquidationID(ctx, row.LiquidationID); t != nil {
			taskSnap = *t
		}
	}
	mid := taskSnap.TriggerMarkPrice
	if taskSnap.LastMidPrice != nil && taskSnap.LastMidPrice.IsPositive() {
		mid = *taskSnap.LastMidPrice
	}
	boundary := mid.Mul(decimal.NewFromInt(1).Sub(HardBoundaryPct(strategy, taskSnap.TriggerCollateralAsset)))

	for _, t := range trades {
		price, _ := decimal.NewFromString(t.Price)
		qty, _ := decimal.NewFromString(t.Qty)
		amount, _ := decimal.NewFromString(t.QuoteQty)
		fee, _ := decimal.NewFromString(t.Commission)
		raw, _ := json.Marshal(t)
		isMaker := t.IsMaker
		feeAsset := t.CommissionAsset
		fill := &entity.LiquidationTrader{
			LiquidationID:      row.LiquidationID,
			BatchNo:            row.BatchNo,
			LiquidationOrderID: row.ID,
			ExchangeCode:       entity.ExchangeCodeBinance,
			ExchangeTradeID:    strconv.FormatInt(t.ID, 10),
			Symbol:             row.Symbol,
			Price:              price,
			Quantity:           qty,
			Amount:             amount,
			Fee:                fee,
			FeeAsset:           &feeAsset,
			IsMaker:            &isMaker,
			TradeTime:          time.UnixMilli(t.Time),
			RawPayload:         string(raw),
		}
		inserted, err := s.fillRepo.CreateIgnoreDuplicate(ctx, tx, fill)
		if err != nil {
			return err
		}
		if inserted && price.LessThanOrEqual(boundary) {
			AlertLiquidation(ctx, "hard_boundary", "成交价触及硬边界", map[string]string{
				"liquidation_id": row.LiquidationID,
				"price":          price.String(),
				"boundary":       boundary.String(),
			})
			if taskSnap.LiquidationID != "" && taskSnap.Status == entity.LiquidationStatusLiquidating {
				return s.terminalHardBoundary(ctx, tx, taskSnap)
			}
		}
	}
	return nil
}

// terminalHardBoundary 硬边界 breached 终态。
func (s *BinanceSync) terminalHardBoundary(ctx context.Context, tx *gorm.DB, task entity.LiquidationTask) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      entity.LiquidationStatusFailed,
		"end_reason":  entity.EndReasonHardBoundaryBreached,
		"finished_at": now,
	}
	if strings.TrimSpace(task.CallbackURL) != "" {
		updates["callback_status"] = entity.CallbackStatusPending
	}
	db := s.db
	if tx != nil {
		db = tx
	}
	if err := s.taskRepo.UpdateByLiquidationID(ctx, db, task.LiquidationID, updates); err != nil {
		return err
	}
	EmitTaskEvent(ctx, db, s.eventRepo, task.LiquidationID, entity.TaskEventAlert, task.Status, entity.LiquidationStatusFailed, "hard boundary breached", nil)
	return nil
}

// IsInsufficientBalance 是否 Binance 余额不足（含 -2010）。
func IsInsufficientBalance(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "insufficient balance") || strings.Contains(msg, "-2010")
}
