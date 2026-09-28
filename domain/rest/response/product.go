package response

type ProductListResp struct {
	Products []ProductItemResp `json:"products"`
}

type ProductItemResp struct {
	ID                   uint     `json:"id"`
	SKUCode              string   `json:"sku_code"`
	CardTitle            string   `json:"card_title"`
	CardSubtitle         string   `json:"card_subtitle"`
	CardFeatures         []string `json:"card_features"`
	ShareSeats           int      `json:"share_seats"`
	SKUUpstreamName      string   `json:"sku_upstream_name"`
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
	IsHot                bool     `json:"is_hot"`
	IsAPIEnabled         bool     `json:"is_api_enabled"`
	TopupTokenAmount     *int64   `json:"topup_token_amount,omitempty"`
	SortOrder            int      `json:"sort_order"`
	Status               string   `json:"status"`
	Flagship             bool     `json:"flagship,omitempty"`
}
