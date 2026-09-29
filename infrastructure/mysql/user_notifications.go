package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserNotificationsImpl struct {
	db *gorm.DB
}

func NewUserNotificationsImpl(db *gorm.DB) repository.UserNotificationsRepo {
	return &UserNotificationsImpl{db: db}
}

func (r *UserNotificationsImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.UserNotifications) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}
