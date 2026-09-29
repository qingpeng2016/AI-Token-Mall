package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserInvoicesImpl struct {
	db *gorm.DB
}

func NewUserInvoicesImpl(db *gorm.DB) repository.UserInvoicesRepo {
	return &UserInvoicesImpl{db: db}
}

func (r *UserInvoicesImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.UserInvoices) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}
