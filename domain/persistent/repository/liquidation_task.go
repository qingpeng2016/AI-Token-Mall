package repository

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"gorm.io/gorm"
)

type LiquidationTaskRepo interface {
	Create(ctx context.Context, tx *gorm.DB, task *entity.LiquidationTask) error
	FindByLiquidationID(ctx context.Context, liquidationID string) (*entity.LiquidationTask, error)
	FindActiveByLoanOrderID(ctx context.Context, loanOrderID string) (*entity.LiquidationTask, error)
	UpdateCancel(ctx context.Context, tx *gorm.DB, liquidationID string, updates map[string]interface{}) error
	UpdateByLiquidationID(ctx context.Context, tx *gorm.DB, liquidationID string, updates map[string]interface{}) error
	// ListRunnable limit<=0 表示不限制条数
	ListRunnable(ctx context.Context, limit int) ([]entity.LiquidationTask, error)
	ListPausedOrCancelled(ctx context.Context) ([]entity.LiquidationTask, error)
	ListPendingCallback(ctx context.Context, limit int) ([]entity.LiquidationTask, error)
	CountLiquidatingBySymbol(ctx context.Context, symbol string) (int64, error)
}
