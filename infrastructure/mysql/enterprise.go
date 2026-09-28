package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type EnterpriseImpl struct {
	db *gorm.DB
}

func NewEnterpriseImpl(db *gorm.DB) repository.EnterpriseRepo {
	return &EnterpriseImpl{db: db}
}

func (r *EnterpriseImpl) CreateInquiry(ctx context.Context, row *entity.EnterpriseInquiry) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *EnterpriseImpl) ListActiveProducts(ctx context.Context) ([]entity.EnterpriseProduct, error) {
	var rows []entity.EnterpriseProduct
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *EnterpriseImpl) FindActiveProductByCode(ctx context.Context, code string) (*entity.EnterpriseProduct, error) {
	var row entity.EnterpriseProduct
	err := r.db.WithContext(ctx).
		Where("code = ? AND status = ?", code, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
