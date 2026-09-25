package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserRepo interface {
	Create(ctx context.Context, tx *gorm.DB, u *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByPhone(ctx context.Context, phone string) (*entity.User, error)
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	UpdateLastLogin(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
}

type StatsRepo interface {
	CountUsers(ctx context.Context) (int64, error)
	CountAccessLogs(ctx context.Context) (int64, error)
}
