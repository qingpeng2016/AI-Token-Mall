package repository

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LiquidationOrderRepo interface {
	Create(ctx context.Context, tx *gorm.DB, row *entity.LiquidationOrder) error
	Update(ctx context.Context, tx *gorm.DB, id uint, updates map[string]interface{}) error
	FindByClientOrderID(ctx context.Context, clientOrderID string) (*entity.LiquidationOrder, error)
	HasInFlightByLiquidationID(ctx context.Context, liquidationID string) (bool, error)
	ListReconcileCandidates(ctx context.Context, limit int) ([]entity.LiquidationOrder, error)
	ListByLiquidationID(ctx context.Context, liquidationID string) ([]entity.LiquidationOrder, error)
	CountByLiquidationID(ctx context.Context, liquidationID string) (int64, error)
	CountConsecutiveIOCUnfilled(ctx context.Context, liquidationID string) (int, error)
	SumInFlightSellQtyBySymbol(ctx context.Context, symbol string) (decimal.Decimal, error)
	ListInFlightByLiquidationID(ctx context.Context, liquidationID string) ([]entity.LiquidationOrder, error)
	// NextBatchSeq 下一卖出轮次序号（按已有子单数量 +1，用于 batch_no / client_order_id）。
	NextBatchSeq(ctx context.Context, liquidationID string) (uint, error)
}
