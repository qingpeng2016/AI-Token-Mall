package repository

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"gorm.io/gorm"
)

type LiquidationCallbackLogRepo interface {
	Create(ctx context.Context, tx *gorm.DB, row *entity.LiquidationCallbackLog) error
	Update(ctx context.Context, tx *gorm.DB, id uint, updates map[string]interface{}) error
	LatestByLiquidationID(ctx context.Context, liquidationID string) (*entity.LiquidationCallbackLog, error)
}
