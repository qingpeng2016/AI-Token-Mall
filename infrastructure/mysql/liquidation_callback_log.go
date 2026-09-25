package mysql

import (
	"context"
	"errors"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"gorm.io/gorm"
)

type LiquidationCallbackLogImpl struct {
	db *gorm.DB
}

func NewLiquidationCallbackLogImpl(db *gorm.DB) repository.LiquidationCallbackLogRepo {
	return &LiquidationCallbackLogImpl{db: db}
}

func (r *LiquidationCallbackLogImpl) Create(ctx context.Context, tx *gorm.DB, row *entity.LiquidationCallbackLog) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(row).Error
}

func (r *LiquidationCallbackLogImpl) Update(ctx context.Context, tx *gorm.DB, id uint, updates map[string]interface{}) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Model(&entity.LiquidationCallbackLog{}).Where("id = ?", id).Updates(updates).Error
}

func (r *LiquidationCallbackLogImpl) LatestByLiquidationID(ctx context.Context, liquidationID string) (*entity.LiquidationCallbackLog, error) {
	var row entity.LiquidationCallbackLog
	err := r.db.WithContext(ctx).
		Where("liquidation_id = ?", liquidationID).
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
