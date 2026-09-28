package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type ProductRepo interface {
	ListOnSale(ctx context.Context, upstreamName string) ([]entity.Product, error)
}
