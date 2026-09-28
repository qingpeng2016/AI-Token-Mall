package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type ProductImpl struct {
	db *gorm.DB
}

func NewProductImpl(db *gorm.DB) repository.ProductRepo {
	return &ProductImpl{db: db}
}

func (r *ProductImpl) ListActiveCategories(ctx context.Context) ([]entity.ProductCategory, error) {
	var rows []entity.ProductCategory
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *ProductImpl) ListOnSale(ctx context.Context, categoryID uint) ([]entity.Product, error) {
	q := r.db.WithContext(ctx).Where("status = ?", "on_sale")
	if categoryID > 0 {
		q = q.Where("products_category_id = ?", categoryID)
	}
	var rows []entity.Product
	err := q.Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}
