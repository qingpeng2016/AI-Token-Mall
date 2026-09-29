package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserWalletFlowsRepo interface {
	SumBalance(ctx context.Context, tx *gorm.DB, userID uint) (decimal.Decimal, error)
	CreateFlow(ctx context.Context, tx *gorm.DB, m *entity.UserWalletFlows) error
}
