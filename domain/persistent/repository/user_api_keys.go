package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserAPIKeysRepo interface {
	Create(ctx context.Context, tx *gorm.DB, m *entity.UserAPIKeys) error
	UpdateWhere(ctx context.Context, tx *gorm.DB, where map[string]interface{}, updateData map[string]interface{}) error
}
