package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserInvoicesRepo interface {
	Create(ctx context.Context, tx *gorm.DB, m *entity.UserInvoices) error
}
