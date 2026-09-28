package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type EnterpriseRepo interface {
	CreateInquiry(ctx context.Context, row *entity.EnterpriseInquiry) error
	ListActiveProducts(ctx context.Context) ([]entity.EnterpriseProduct, error)
	FindActiveProductByCode(ctx context.Context, code string) (*entity.EnterpriseProduct, error)
}
