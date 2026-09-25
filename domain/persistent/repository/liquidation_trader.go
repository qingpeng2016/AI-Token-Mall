package repository

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LiquidationTraderRepo interface {
	CreateIgnoreDuplicate(ctx context.Context, tx *gorm.DB, row *entity.LiquidationTrader) (inserted bool, err error)
	AggregateByLiquidationID(ctx context.Context, liquidationID string) (qty, amount, fee decimal.Decimal, err error)
	ListByLiquidationID(ctx context.Context, liquidationID string) ([]entity.LiquidationTrader, error)
	AggregateByBatchNo(ctx context.Context, batchNo string) (qty, amount, fee decimal.Decimal, err error)
}
