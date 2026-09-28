package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type ProductRepo interface {
	ListActiveCategories(ctx context.Context) ([]entity.ProductCategory, error)
	ListOnSale(ctx context.Context, categoryID uint) ([]entity.Product, error)
	FindOnSaleBySKUCode(ctx context.Context, skuCode string) (*entity.Product, error)
}
