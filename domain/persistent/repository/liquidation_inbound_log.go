package repository

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"gorm.io/gorm"
)

type LiquidationInboundLogRepo interface {
	Create(ctx context.Context, tx *gorm.DB, row *entity.LiquidationInboundLog) error
	UpdateResponse(ctx context.Context, tx *gorm.DB, id uint, httpStatus int, responseBody string) error
	ListByLiquidationID(ctx context.Context, liquidationID string, limit int) ([]entity.LiquidationInboundLog, error)
}
