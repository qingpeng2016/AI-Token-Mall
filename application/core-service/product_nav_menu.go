package coreservice

import (
	"fmt"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

var skuToSlug = map[string]string{
	"OAI-GO-M":     "gpt-go",
	"OAI-PLUS-M":   "chatgpt-plus",
	"OAI-PRO5-M":   "chatgpt-pro-5x",
	"OAI-PRO20-M":  "chatgpt-pro-20x",
	"ANT-PRO-M":    "claude-pro",
	"ANT-MAX-M":    "claude-max-5x",
	"XAI-GROK-M":   "grok-super",
	"GEM-PRO-M":    "gemini-pro",
	"CUR-PRO-M":    "cursor-pro",
	"PPX-PRO-M":    "perplexity-pro",
}

var navColumnOrder = []string{"openai", "anthropic", "xai", "gemini", "cursor", "perplexity"}

var navColumnTitle = map[string]string{
	"openai":      "ChatGPT",
	"anthropic":   "Claude",
	"xai":         "Grok",
	"gemini":      "Gemini",
	"cursor":      "Cursor",
	"perplexity":  "Perplexity",
}

var brandMenuFromColumn = []struct {
	id      string
	column  string
	label   string
	hotSale bool
}{
	{id: "chatgpt", column: "openai", label: "ChatGPT"},
	{id: "claude", column: "anthropic", label: "Claude"},
	{id: "cursor", column: "cursor", label: "Cursor"},
}

func (s *ProductService) buildNavMenu(products []entity.Product) *response.NavMenuResp {
	byColumn := map[string][]response.NavPlanItemResp{}
	for i := range products {
		p := &products[i]
		key := navColumnKey(p)
		item := mapProductToNavItem(p)
		byColumn[key] = append(byColumn[key], item)
	}

	mega := make([]response.NavMegaColumnResp, 0, len(navColumnOrder))
	for _, key := range navColumnOrder {
		items := byColumn[key]
		if len(items) == 0 {
			continue
		}
		title := navColumnTitle[key]
		if title == "" {
			title = key
		}
		mega = append(mega, response.NavMegaColumnResp{Title: title, Items: items})
	}

	brands := make([]response.NavBrandMenuResp, 0, len(brandMenuFromColumn))
	for _, bm := range brandMenuFromColumn {
		items := byColumn[bm.column]
		if len(items) == 0 {
			continue
		}
		brands = append(brands, response.NavBrandMenuResp{
			ID:      bm.id,
			Label:   bm.label,
			HotSale: bm.hotSale,
			Items:   items,
		})
	}

	return &response.NavMenuResp{MegaMenu: mega, BrandMenus: brands}
}

func navColumnKey(p *entity.Product) string {
	if strings.HasPrefix(p.SKUCode, "CUR-") {
		return "cursor"
	}
	return p.SKUUpstreamName
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
