package coreservice

import (
	"context"
	"encoding/json"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

type ProductService struct {
	productRepo repository.ProductRepo
}

func NewProductService(productRepo repository.ProductRepo) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) List(ctx context.Context, upstreamName string) (*response.ProductListResp, error) {
	rows, err := s.productRepo.ListOnSale(ctx, upstreamName)
	if err != nil {
		return nil, err
	}
	items := make([]response.ProductItemResp, 0, len(rows))
	for i := range rows {
		items = append(items, mapProductItem(&rows[i]))
	}
	return &response.ProductListResp{Products: items}, nil
}

func mapProductItem(p *entity.Product) response.ProductItemResp {
	item := response.ProductItemResp{
		ID:              p.ID,
		SKUCode:         p.SKUCode,
		CardTitle:       p.CardTitle,
		CardSubtitle:    p.CardSubtitle,
		CardFeatures: expandCardFeatures(decodeStringJSONArray(p.CardFeatures), p),
		ShareSeats:      p.ShareSeats,
		SKUUpstreamName: p.SKUUpstreamName,
		SKUProductName:  p.SKUProductName,
		LimitTokens:     p.LimitTokens,
		RPMLimit:        p.RPMLimit,
		TPMLimit:        p.TPMLimit,
		AllowedModels:   decodeStringJSONArray(p.AllowedModels),
		ProductType:     p.ProductType,
		BillingPeriod:   p.BillingPeriod,
		PriceCents:      p.PriceCents,
		Currency:        p.Currency,
		CompareAtPriceCents: p.CompareAtPriceCents,
		Highlights:      decodeStringJSONArray(p.HighlightsJSON),
		IsHot:           p.IsHot != 0,
		IsAPIEnabled:    p.IsAPIEnabled != 0,
		TopupTokenAmount: p.TopupTokenAmount,
		SortOrder:       p.SortOrder,
		Status:          p.Status,
		Flagship:        p.IsFlagship != 0,
	}
	if len(item.Highlights) == 0 {
		item.Highlights = []string{}
	}
	if len(item.CardFeatures) == 0 {
		item.CardFeatures = []string{}
	}
	if len(item.AllowedModels) == 0 {
		item.AllowedModels = []string{}
	}
	return item
}

func decodeStringJSONArray(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}
