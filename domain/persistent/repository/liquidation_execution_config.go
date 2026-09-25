package repository

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
)

type LiquidationExecutionConfigRepo interface {
	GetStrategy(ctx context.Context) (*entity.LiquidationStrategyConfig, error)
	ListTiers(ctx context.Context) ([]entity.LiquidationTierConfig, error)
}
