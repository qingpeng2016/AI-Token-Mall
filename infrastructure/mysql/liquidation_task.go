package mysql

import (
	"context"
	"errors"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"gorm.io/gorm"
)

type LiquidationTaskImpl struct {
	db *gorm.DB
}

func NewLiquidationTaskImpl(db *gorm.DB) repository.LiquidationTaskRepo {
	return &LiquidationTaskImpl{db: db}
}

func (r *LiquidationTaskImpl) Create(ctx context.Context, tx *gorm.DB, task *entity.LiquidationTask) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(task).Error
}

func (r *LiquidationTaskImpl) FindByLiquidationID(ctx context.Context, liquidationID string) (*entity.LiquidationTask, error) {
	var task entity.LiquidationTask
	err := r.db.WithContext(ctx).Where("liquidation_id = ?", liquidationID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *LiquidationTaskImpl) FindActiveByLoanOrderID(ctx context.Context, loanOrderID string) (*entity.LiquidationTask, error) {
	var task entity.LiquidationTask
	err := r.db.WithContext(ctx).
		Where("loan_order_id = ?", loanOrderID).
		Where("status = ?", entity.LiquidationStatusLiquidating).
		Order("id DESC").
		First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *LiquidationTaskImpl) UpdateCancel(ctx context.Context, tx *gorm.DB, liquidationID string, updates map[string]interface{}) error {
	return r.UpdateByLiquidationID(ctx, tx, liquidationID, updates)
}

func (r *LiquidationTaskImpl) UpdateByLiquidationID(ctx context.Context, tx *gorm.DB, liquidationID string, updates map[string]interface{}) error {
	db = r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Model(&entity.LiquidationTask{}).
		Where("liquidation_id = ?", liquidationID).
		Updates(updates).Error
}

func (r *LiquidationTaskImpl) ListRunnable(ctx context.Context, limit int) ([]entity.LiquidationTask, error) {
	var rows []entity.LiquidationTask
	q := r.db.WithContext(ctx).
		Where("status = ?", entity.LiquidationStatusLiquidating).
		Order("updated_at ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *LiquidationTaskImpl) ListPausedOrCancelled(ctx context.Context) ([]entity.LiquidationTask, error) {
	var rows []entity.LiquidationTask
	err := r.db.WithContext(ctx).
		Where("status IN ?", []string{entity.LiquidationStatusPaused, entity.LiquidationStatusCancelled}).
		Where("in_flight_orders_cleared = ?", entity.InFlightOrdersClearedNo).
		Order("updated_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *LiquidationTaskImpl) ListPendingCallback(ctx context.Context, limit int) ([]entity.LiquidationTask, error) {
	if limit <= 0 {
		limit = 20
	}
	terminal := []string{
		entity.LiquidationStatusSettled,
		entity.LiquidationStatusBadDebt,
		entity.LiquidationStatusCancelled,
		entity.LiquidationStatusFailed,
	}
	var rows []entity.LiquidationTask
	err := r.db.WithContext(ctx).
		Where("status IN ?", terminal).
		Where("(callback_status IS NULL OR callback_status = ?)", entity.CallbackStatusPending).
		Order("updated_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *LiquidationTaskImpl) CountLiquidatingBySymbol(ctx context.Context, symbol string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.LiquidationTask{}).
		Where("symbol = ?", symbol).
		Where("status = ?", entity.LiquidationStatusLiquidating).
		Count(&n).Error
	return n, err
}
