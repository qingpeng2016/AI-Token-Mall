package coreservice

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

var ErrProductNotFound = errors.New("product not found")

type ProductService struct {
	productRepo repository.ProductRepo
}

func NewProductService(productRepo repository.ProductRepo) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) List(ctx context.Context, categoryID uint) (*response.ProductCatalogResp, error) {
	categories, err := s.productRepo.ListActiveCategories(ctx)
	if err != nil {
		return nil, err
	}
	products, err := s.productRepo.ListOnSale(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	return buildProductCatalog(categories, products), nil
}

func (s *ProductService) NavMenu(ctx context.Context) (*response.NavMenuResp, error) {
	categories, err := s.productRepo.ListActiveCategories(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.productRepo.ListOnSale(ctx, 0)
	if err != nil {
		return nil, err
	}
	return s.buildNavMenu(categories, rows), nil
}

func buildProductCatalog(categories []entity.ProductCategory, products []entity.Product) *response.ProductCatalogResp {
	byCategoryID := map[uint][]response.ProductItemResp{}
	for i := range products {
		p := &products[i]
		if p.ProductsCategoryID == nil {
			continue
		}
		id := *p.ProductsCategoryID
		byCategoryID[id] = append(byCategoryID[id], mapProductItem(p))
	}

	groups := make([]response.ProductCategoryGroupResp, 0, len(categories))
	for i := range categories {
		c := &categories[i]
		items := byCategoryID[c.ID]
		if items == nil {
			items = []response.ProductItemResp{}
		}
		groups = append(groups, response.ProductCategoryGroupResp{
			ID:         c.ID,
			Name:       c.Name,
			DotColor:   c.DotColor,
			ActiveBg:   c.ActiveBg,
			Sort:       c.Sort,
			HotTagName: strings.TrimSpace(c.HotTagName),
			Products:   items,
		})
	}
	return &response.ProductCatalogResp{Categories: groups}
}

func mapProductItem(p *entity.Product) response.ProductItemResp {
	item := response.ProductItemResp{
		ID:                   p.ID,
		SKUCode:              p.SKUCode,
		CardTitle:            p.CardTitle,
		CardSubtitle:         p.CardSubtitle,
		CardFeatures:         expandCardFeatures(decodeStringJSONArray(p.CardFeatures), p),
		ShareSeats:           p.ShareSeats,
		ProductsCategoryID:   p.ProductsCategoryID,
		ProductsCategoryName: p.ProductsCategoryName,
		SKUProductName:       p.SKUProductName,
		LimitTokens:          p.LimitTokens,
		RPMLimit:             p.RPMLimit,
		TPMLimit:             p.TPMLimit,
		AllowedModels:        decodeStringJSONArray(p.AllowedModels),
		ProductType:          p.ProductType,
		BillingPeriod:        p.BillingPeriod,
		PriceCents:           p.PriceCents,
		Currency:             p.Currency,
		CompareAtPriceCents:  p.CompareAtPriceCents,
		Highlights:           decodeStringJSONArray(p.HighlightsJSON),
		HotTagName:           strings.TrimSpace(p.HotTagName),
		IsAPIEnabled:         p.IsAPIEnabled != 0,
		TopupTokenAmount:     p.TopupTokenAmount,
		Sort:                 p.SortOrder,
		Status:               p.Status,
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

func (s *ProductService) DetailBySlug(ctx context.Context, slug string) (*response.ProductDetailResp, error) {
	sku := skuFromSlug(slug)
	if sku == "" {
		return nil, ErrProductNotFound
	}
	p, err := s.productRepo.FindOnSaleBySKUCode(ctx, sku)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	item := mapProductItem(p)
	detail := buildProductDetailContent(slug, p, item)
	return &response.ProductDetailResp{
		Slug:    canonicalSlugForSKU(p.SKUCode),
		Product: item,
		Detail:  detail,
	}, nil
}
