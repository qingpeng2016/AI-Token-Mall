package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserWalletFlowsImpl struct {
	db *gorm.DB
}

func NewUserWalletFlowsImpl(db *gorm.DB) repository.UserWalletFlowsRepo {
	return &UserWalletFlowsImpl{db: db}
}

func (r *UserWalletFlowsImpl) SumBalance(ctx context.Context, tx *gorm.DB, userID uint) (decimal.Decimal, error) {
	var sum decimal.Decimal
	err := repository.GormDB(ctx, r.db, tx).Model(&entity.UserWalletFlows{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *UserWalletFlowsImpl) CreateFlow(ctx context.Context, tx *gorm.DB, m *entity.UserWalletFlows) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}
