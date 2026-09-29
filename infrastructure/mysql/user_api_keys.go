package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserAPIKeysImpl struct {
	db *gorm.DB
}

func NewUserAPIKeysImpl(db *gorm.DB) repository.UserAPIKeysRepo {
	return &UserAPIKeysImpl{db: db}
}

func (r *UserAPIKeysImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.UserAPIKeys) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}

func (r *UserAPIKeysImpl) UpdateWhere(ctx context.Context, tx *gorm.DB, where map[string]interface{}, updateData map[string]interface{}) error {
	db := repository.GormDB(ctx, r.db, tx).Model(entity.UserAPIKeys{})
	for condition, value := range where {
		db = db.Where(condition, value)
	}
	return db.Updates(updateData).Error
}
