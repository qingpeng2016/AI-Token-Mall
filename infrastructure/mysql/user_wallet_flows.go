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

func (r *UserWalletFlowsImpl) ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.UserWalletFlows, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var rows []entity.UserWalletFlows
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
