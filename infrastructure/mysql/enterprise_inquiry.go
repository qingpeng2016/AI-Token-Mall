package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type EnterpriseInquiryImpl struct {
	db *gorm.DB
}

func NewEnterpriseInquiryImpl(db *gorm.DB) repository.EnterpriseInquiryRepo {
	return &EnterpriseInquiryImpl{db: db}
}

func (r *EnterpriseInquiryImpl) Create(ctx context.Context, row *entity.EnterpriseInquiry) error {
	return r.db.WithContext(ctx).Create(row).Error
}
