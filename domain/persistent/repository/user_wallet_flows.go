package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserWalletFlowsRepo interface {
	CreateFlow(ctx context.Context, tx *gorm.DB, m *entity.UserWalletFlows) error
	ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.UserWalletFlows, error)
}
