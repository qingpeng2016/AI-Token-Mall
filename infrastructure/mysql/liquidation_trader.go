package mysql

import (
	"context"
	"errors"
	"strings"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LiquidationTraderImpl struct {
	db *gorm.DB
}

func NewLiquidationTraderImpl(db *gorm.DB) repository.LiquidationTraderRepo {
	return &LiquidationTraderImpl{db: db}
}

func (r *LiquidationTraderImpl) CreateIgnoreDuplicate(ctx context.Context, tx *gorm.DB, row *entity.LiquidationTrader) (bool, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	err := db.WithContext(ctx).Create(row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return false, nil
		}
		if isMySQLDuplicate(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *LiquidationTraderImpl) AggregateByLiquidationID(ctx context.Context, liquidationID string) (qty, amount, fee decimal.Decimal, err error) {
	type agg struct {
		Qty    decimal.Decimal
		Amount decimal.Decimal
		Fee    decimal.Decimal
	}
	var a agg
	err = r.db.WithContext(ctx).Model(&entity.LiquidationTrader{}).
		Select("COALESCE(SUM(quantity),0) as qty, COALESCE(SUM(amount),0) as amount, COALESCE(SUM(fee),0) as fee").
		Where("liquidation_id = ?", liquidationID).
		Scan(&a).Error
	return a.Qty, a.Amount, a.Fee, err
}

func (r *LiquidationTraderImpl) ListByLiquidationID(ctx context.Context, liquidationID string) ([]entity.LiquidationTrader, error) {
	var rows []entity.LiquidationTrader
	err := r.db.WithContext(ctx).
		Where("liquidation_id = ?", liquidationID).
		Order("trade_time ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *LiquidationTraderImpl) AggregateByBatchNo(ctx context.Context, batchNo string) (qty, amount, fee decimal.Decimal, err error) {
	type agg struct {
		Qty    decimal.Decimal
		Amount decimal.Decimal
		Fee    decimal.Decimal
	}
	var a agg
	err = r.db.WithContext(ctx).Model(&entity.LiquidationTrader{}).
		Select("COALESCE(SUM(quantity),0) as qty, COALESCE(SUM(amount),0) as amount, COALESCE(SUM(fee),0) as fee").
		Where("batch_no = ?", batchNo).
		Scan(&a).Error
	return a.Qty, a.Amount, a.Fee, err
}

func isMySQLDuplicate(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "1062")
}
