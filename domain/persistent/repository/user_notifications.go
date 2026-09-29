package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserNotificationsRepo interface {
	Create(ctx context.Context, tx *gorm.DB, m *entity.UserNotifications) error
}
