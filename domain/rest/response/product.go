package response

type ProductCatalogResp struct {
	Categories []ProductCategoryGroupResp `json:"categories"`
}

type ProductCategoryGroupResp struct {
	ID       uint              `json:"id"`
	Name     string            `json:"name"`
	DotColor string            `json:"dot_color,omitempty"`
	ActiveBg             string            `json:"active_bg,omitempty"`
	Sort                 int               `json:"sort"`
	HotTagName           string            `json:"hot_tag_name,omitempty"`
	Products             []ProductItemResp `json:"products"`
}

type ProductItemResp struct {
	ID                   uint     `json:"id"`
	SKUCode              string   `json:"sku_code"`
	CardTitle            string   `json:"card_title"`
	CardSubtitle         string   `json:"card_subtitle"`
	CardFeatures         []string `json:"card_features"`
	ShareSeats           int      `json:"share_seats"`
	ProductsCategoryID   *uint    `json:"products_category_id,omitempty"`
	ProductsCategoryName string   `json:"products_category_name"`
	SKUProductName       string   `json:"sku_product_name"`
	LimitTokens          int64    `json:"limit_tokens"`
	RPMLimit             int      `json:"rpm_limit"`
	TPMLimit             *int     `json:"tpm_limit,omitempty"`
	AllowedModels        []string `json:"allowed_models"`
	ProductType          string   `json:"product_type"`
	BillingPeriod        string   `json:"billing_period"`
	PriceCents           int64    `json:"price_cents"`
	Currency             string   `json:"currency"`
	CompareAtPriceCents  *int64   `json:"compare_at_price_cents,omitempty"`
	Highlights           []string `json:"highlights"`
	HotTagName           string   `json:"hot_tag_name,omitempty"`
	IsAPIEnabled         bool     `json:"is_api_enabled"`
	TopupTokenAmount     *int64   `json:"topup_token_amount,omitempty"`
	Sort                 int      `json:"sort"`
	Status               string   `json:"status"`
}
