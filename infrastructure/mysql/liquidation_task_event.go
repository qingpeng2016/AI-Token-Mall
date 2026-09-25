package mysql

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"gorm.io/gorm"
)

type LiquidationTaskEventImpl struct {
	db *gorm.DB
}

func NewLiquidationTaskEventImpl(db *gorm.DB) repository.LiquidationTaskEventRepo {
	return &LiquidationTaskEventImpl{db: db}
}

func (r *LiquidationTaskEventImpl) Create(ctx context.Context, tx *gorm.DB, row *entity.LiquidationTaskEvent) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(row).Error
}
