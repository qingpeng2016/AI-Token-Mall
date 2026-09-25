package repository

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"gorm.io/gorm"
)

type LiquidationTaskEventRepo interface {
	Create(ctx context.Context, tx *gorm.DB, row *entity.LiquidationTaskEvent) error
}
