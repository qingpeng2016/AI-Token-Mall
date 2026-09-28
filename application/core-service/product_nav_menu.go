package coreservice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

const navTopBrandCategoryLimit = 3

var skuToSlug = map[string]string{
	"OAI-GO-M":    "gpt-go",
	"OAI-PLUS-M":  "chatgpt-plus",
	"OAI-PRO5-M":  "chatgpt-pro-5x",
	"OAI-PRO20-M": "chatgpt-pro-20x",
	"ANT-PRO-M":   "claude-pro",
	"ANT-MAX-M":   "claude-max-5x",
	"XAI-GROK-M":  "grok-super",
	"GEM-PRO-M":   "gemini-pro",
	"CUR-PRO-M":   "cursor-pro",
	"PPX-PRO-M":   "perplexity-pro",
}

func (s *ProductService) buildNavMenu(categories []entity.ProductCategory, products []entity.Product) *response.NavMenuResp {
	byCategoryID := map[uint][]response.NavPlanItemResp{}
	for i := range products {
		p := &products[i]
		if p.ProductsCategoryID == nil {
			continue
		}
		id := *p.ProductsCategoryID
		byCategoryID[id] = append(byCategoryID[id], mapProductToNavItem(p))
	}

	mega := make([]response.NavMegaColumnResp, 0, len(categories))
	brands := make([]response.NavBrandMenuResp, 0, navTopBrandCategoryLimit)

	for _, cat := range categories {
		items := byCategoryID[cat.ID]
		if len(items) == 0 {
			continue
		}
		mega = append(mega, response.NavMegaColumnResp{Title: cat.Name, Items: items})
		if len(brands) < navTopBrandCategoryLimit {
			brands = append(brands, response.NavBrandMenuResp{
				ID:         strconv.FormatUint(uint64(cat.ID), 10),
				Label:      cat.Name,
				HotTagName: strings.TrimSpace(cat.HotTagName),
				Items:      items,
			})
		}
	}

	return &response.NavMenuResp{MegaMenu: mega, BrandMenus: brands}
}

func mapProductToNavItem(p *entity.Product) response.NavPlanItemResp {
	label := strings.TrimSpace(p.CardTitle)
	label = strings.TrimSuffix(label, "月卡")
	label = strings.TrimSpace(label)
	if label == "" {
		label = p.SKUProductName
	}
	return response.NavPlanItemResp{
		Label: label,
		Price: formatNavPrice(p.PriceCents, p.BillingPeriod),
		Href:  productDetailHref(p.SKUCode),
	}
}

func productDetailHref(skuCode string) string {
	if slug, ok := skuToSlug[skuCode]; ok {
		return "/p/" + slug
	}
	return "/p/" + strings.ToLower(strings.ReplaceAll(skuCode, "_", "-"))
}

func formatNavPrice(cents int64, billingPeriod string) string {
	yuan := formatCnyFromCents(cents)
	switch billingPeriod {
	case "month":
		return yuan + "/月"
	case "year":
		return yuan + "/年"
	case "once":
		return yuan
	default:
		return yuan + "/月"
	}
}

func formatCnyFromCents(cents int64) string {
	if cents%100 == 0 {
		return fmt.Sprintf("¥%d", cents/100)
	}
	return fmt.Sprintf("¥%.2f", float64(cents)/100)
}
