package mysql

import (
	"context"
	"errors"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"gorm.io/gorm"
)

type LiquidationExecutionConfigImpl struct {
	db *gorm.DB
}

func NewLiquidationExecutionConfigImpl(db *gorm.DB) repository.LiquidationExecutionConfigRepo {
	return &LiquidationExecutionConfigImpl{db: db}
}

func (r *LiquidationExecutionConfigImpl) GetStrategy(ctx context.Context) (*entity.LiquidationStrategyConfig, error) {
	var row entity.LiquidationStrategyConfig
	err := r.db.WithContext(ctx).Where("id = ?", 1).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return defaultStrategyConfig(), nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *LiquidationExecutionConfigImpl) ListTiers(ctx context.Context) ([]entity.LiquidationTierConfig, error) {
	var rows []entity.LiquidationTierConfig
	err := r.db.WithContext(ctx).Where("is_enabled = ?", 1).Order("tier_rank ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return defaultTierConfigs(), nil
	}
	return rows, nil
}

func defaultStrategyConfig() *entity.LiquidationStrategyConfig {
	return &entity.LiquidationStrategyConfig{ID: 1}
}

func defaultTierConfigs() []entity.LiquidationTierConfig {
	// 与 docs/liquidation_init.sql 中 tier seed 一致，DB 未建表时兜底
	return []entity.LiquidationTierConfig{}
}
