package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserWalletFlowsImpl struct {
	db *gorm.DB
}

func NewUserWalletFlowsImpl(db *gorm.DB) repository.UserWalletFlowsRepo {
	return &UserWalletFlowsImpl{db: db}
}

func (r *UserWalletFlowsImpl) CreateFlow(ctx context.Context, tx *gorm.DB, m *entity.UserWalletFlows) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}
