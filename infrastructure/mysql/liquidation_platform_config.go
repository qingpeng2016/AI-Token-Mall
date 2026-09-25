package mysql

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"gorm.io/gorm"
)

type LiquidationPlatformConfigImpl struct {
	db *gorm.DB
}

func NewLiquidationPlatformConfigImpl(db *gorm.DB) repository.LiquidationPlatformConfigRepo {
	return &LiquidationPlatformConfigImpl{db: db}
}

func (r *LiquidationPlatformConfigImpl) FindOne(ctx context.Context, tx *gorm.DB, where map[string]interface{}, orderBy string) (entity.LiquidationPlatformConfig, error) {
	if tx == nil {
		tx = r.db
	}
	db := tx.WithContext(ctx).Model(entity.LiquidationPlatformConfig{})
	for condition, value := range where {
		db = db.Where(condition, value)
	}
	if orderBy != "" {
		db = db.Order(orderBy)
	}
	var m entity.LiquidationPlatformConfig
	err := db.First(&m).Error
	return m, err
}
