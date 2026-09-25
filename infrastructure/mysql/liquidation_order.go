package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const reconcileStaleAfter = 30 * time.Second

type LiquidationOrderImpl struct {
	db *gorm.DB
}

func NewLiquidationOrderImpl(db *gorm.DB) repository.LiquidationOrderRepo {
	return &LiquidationOrderImpl{db: db}
}

func (r *LiquidationOrderImpl) Create(ctx context.Context, tx *gorm.DB, row *entity.LiquidationOrder) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(row).Error
}

func (r *LiquidationOrderImpl) Update(ctx context.Context, tx *gorm.DB, id uint, updates map[string]interface{}) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Model(&entity.LiquidationOrder{}).Where("id = ?", id).Updates(updates).Error
}

func (r *LiquidationOrderImpl) FindByClientOrderID(ctx context.Context, clientOrderID string) (*entity.LiquidationOrder, error) {
	var row entity.LiquidationOrder
	err := r.db.WithContext(ctx).Where("client_order_id = ?", clientOrderID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *LiquidationOrderImpl) HasInFlightByLiquidationID(ctx context.Context, liquidationID string) (bool, error) {
	inFlight := []string{
		entity.ExchangeOrderStatusNew,
		entity.ExchangeOrderStatusUnknown,
		entity.ExchangeOrderStatusPartiallyFilled,
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.LiquidationOrder{}).
		Where("liquidation_id = ?", liquidationID).
		Where("status IN ?", inFlight).
		Count(&n).Error
	return n > 0, err
}

func (r *LiquidationOrderImpl) ListReconcileCandidates(ctx context.Context, limit int) ([]entity.LiquidationOrder, error) {
	if limit <= 0 {
		limit = 50
	}
	staleBefore := time.Now().Add(-reconcileStaleAfter)
	var rows []entity.LiquidationOrder
	err := r.db.WithContext(ctx).
		Where(
			"status = ? OR (status = ? AND submitted_at IS NOT NULL AND submitted_at < ?)",
			entity.ExchangeOrderStatusUnknown,
			entity.ExchangeOrderStatusNew,
			staleBefore,
		).
		Order("updated_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *LiquidationOrderImpl) ListByLiquidationID(ctx context.Context, liquidationID string) ([]entity.LiquidationOrder, error) {
	var rows []entity.LiquidationOrder
	err := r.db.WithContext(ctx).Where("liquidation_id = ?", liquidationID).Order("id ASC").Find(&rows).Error
	return rows, err
}

func (r *LiquidationOrderImpl) CountByLiquidationID(ctx context.Context, liquidationID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.LiquidationOrder{}).Where("liquidation_id = ?", liquidationID).Count(&n).Error
	return n, err
}

func (r *LiquidationOrderImpl) NextBatchSeq(ctx context.Context, liquidationID string) (uint, error) {
	n, err := r.CountByLiquidationID(ctx, liquidationID)
	if err != nil {
		return 0, err
	}
	return uint(n) + 1, nil
}

func (r *LiquidationOrderImpl) CountConsecutiveIOCUnfilled(ctx context.Context, liquidationID string) (int, error) {
	var rows []entity.LiquidationOrder
	err := r.db.WithContext(ctx).Where("liquidation_id = ?", liquidationID).Order("id DESC").Limit(10).Find(&rows).Error
	if err != nil {
		return 0, err
	}
	streak := 0
	for _, o := range rows {
		if o.Status == entity.ExchangeOrderStatusCancelled && o.FilledQty.IsZero() {
			streak++
			continue
		}
		break
	}
	return streak, nil
}

func (r *LiquidationOrderImpl) SumInFlightSellQtyBySymbol(ctx context.Context, symbol string) (decimal.Decimal, error) {
	inFlight := []string{
		entity.ExchangeOrderStatusNew,
		entity.ExchangeOrderStatusUnknown,
		entity.ExchangeOrderStatusPartiallyFilled,
	}
	type row struct {
		S decimal.Decimal
	}
	var res row
	err := r.db.WithContext(ctx).Model(&entity.LiquidationOrder{}).
		Select("COALESCE(SUM(quantity - filled_qty), 0) as s").
		Where("symbol = ?", symbol).
		Where("status IN ?", inFlight).
		Scan(&res).Error
	return res.S, err
}

func (r *LiquidationOrderImpl) ListInFlightByLiquidationID(ctx context.Context, liquidationID string) ([]entity.LiquidationOrder, error) {
	inFlight := []string{
		entity.ExchangeOrderStatusNew,
		entity.ExchangeOrderStatusUnknown,
		entity.ExchangeOrderStatusPartiallyFilled,
	}
	var rows []entity.LiquidationOrder
	err := r.db.WithContext(ctx).
		Where("liquidation_id = ?", liquidationID).
		Where("status IN ?", inFlight).
		Find(&rows).Error
	return rows, err
}
