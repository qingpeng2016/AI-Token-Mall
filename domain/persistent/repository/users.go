package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UsersRepo interface {
	Create(ctx context.Context, tx *gorm.DB, u *entity.Users) error
	FindByEmail(ctx context.Context, email string) (*entity.Users, error)
	FindByPhone(ctx context.Context, phone string) (*entity.Users, error)
	FindByID(ctx context.Context, id uint) (*entity.Users, error)
	FindByVipDomain(ctx context.Context, vipDomain string) (*entity.Users, error)
	ListByParentUserID(ctx context.Context, parentUserID uint, offset, limit int) ([]entity.Users, error)
	CountByParentUserID(ctx context.Context, parentUserID uint) (int64, error)
	CountValidInvitesByParentUserID(ctx context.Context, parentUserID uint) (int64, error)
	FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id uint) (*entity.Users, error)
	ApplyWalletDelta(ctx context.Context, tx *gorm.DB, userID uint, delta decimal.Decimal) (balanceAfter decimal.Decimal, err error)
	UpdateLastLogin(ctx context.Context, id uint) error
	UpdatePassword(ctx context.Context, id uint, passwordHash, passwordPlain string) error
	Count(ctx context.Context) (int64, error)
}

type StatsRepo interface {
	CountUsers(ctx context.Context) (int64, error)
	CountAccessLogs(ctx context.Context) (int64, error)
}
