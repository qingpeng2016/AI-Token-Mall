package mysql

import (
	"context"
	"strings"

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

func (r *ProductImpl) ListOnSale(ctx context.Context, upstreamName string) ([]entity.Product, error) {
	q := r.db.WithContext(ctx).Where("status = ?", "on_sale")
	upstreamName = strings.TrimSpace(upstreamName)
	if upstreamName != "" {
		q = q.Where("sku_upstream_name = ?", upstreamName)
	}
	var rows []entity.Product
	err := q.Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}
