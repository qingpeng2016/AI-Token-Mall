package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type VipConfigImpl struct {
	db *gorm.DB
}

func NewVipConfigImpl(db *gorm.DB) repository.VipConfigRepo {
	return &VipConfigImpl{db: db}
}

func (r *VipConfigImpl) FindByID(ctx context.Context, id uint) (*entity.VipConfig, error) {
	var row entity.VipConfig
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *VipConfigImpl) ListEnabled(ctx context.Context) ([]entity.VipConfig, error) {
	var rows []entity.VipConfig
	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&rows).Error
	return rows, err
}
