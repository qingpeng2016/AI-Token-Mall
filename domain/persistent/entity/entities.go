package entity

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	ID           uint       `gorm:"primaryKey;column:id"`
	Email        *string    `gorm:"column:email;size:255"`
	Phone        *string    `gorm:"column:phone;size:32"`
	PasswordHash string     `gorm:"column:password_hash;size:255;not null"`
	PasswordPlain string    `gorm:"column:password_plain;size:255;not null"`
	Nickname     *string    `gorm:"column:nickname;size:64"`
	Status       string     `gorm:"column:status;size:32;not null;default:active"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }

type UserInvoiceConfig struct {
	ID          uint      `gorm:"primaryKey;column:id"`
	UserID      uint      `gorm:"column:user_id;not null"`
	ProfileType string    `gorm:"column:profile_type;size:16;not null;default:enterprise"`
	Title       string    `gorm:"column:title;size:256;not null"`
	TaxNo       *string   `gorm:"column:tax_no;size:64"`
	BankName    *string   `gorm:"column:bank_name;size:128"`
	BankAccount *string   `gorm:"column:bank_account;size:64"`
	Address     *string   `gorm:"column:address;size:512"`
	Phone       *string   `gorm:"column:phone;size:32"`
	IsDefault   int       `gorm:"column:is_default;not null;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (UserInvoiceConfig) TableName() string { return "user_invoice_config" }

type UpstreamInfo struct {
	ID                   uint       `gorm:"primaryKey;column:id"`
	ProductsCategoryName string     `gorm:"column:products_category_name;size:32;not null"`
	SKUProductName       string     `gorm:"column:sku_product_name;size:128;not null"`
	AccountLabel         string     `gorm:"column:account_label;size:128;not null"`
	APIKeyCiphertext     []byte     `gorm:"column:api_key_ciphertext;not null"`
	APISecretCiphertext  []byte     `gorm:"column:api_secret_ciphertext"`
	ProcurementCostNote  *string    `gorm:"column:procurement_cost_note;size:256"`
	ExpiresAt            *time.Time `gorm:"column:expires_at"`
	CapTokens            *int64     `gorm:"column:cap_tokens"`
	UsedTokens           int64      `gorm:"column:used_tokens;not null;default:0"`
	QuotaResetAt         *time.Time `gorm:"column:quota_reset_at"`
	Weight               int        `gorm:"column:weight;not null;default:100"`
	Status               string     `gorm:"column:status;size:32;not null;default:active"`
	FailRate             *float64   `gorm:"column:fail_rate"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (UpstreamInfo) TableName() string { return "upstream_info" }

type ProductCategory struct {
	ID       uint      `gorm:"primaryKey;column:id"`
	Name     string    `gorm:"column:name;size:64;not null"`
	DotColor string    `gorm:"column:dot_color;size:16"`
	ActiveBg             string    `gorm:"column:active_bg;size:16"`
	Sort                 int       `gorm:"column:sort;not null;default:0"`
	Status               string    `gorm:"column:status;size:16;not null;default:active"`
	HotTagName           string    `gorm:"column:hot_tag_name;size:32"`
	CreatedAt            time.Time `gorm:"column:created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

func (ProductCategory) TableName() string { return "products_category" }

type Product struct {
	ID                  uint           `gorm:"primaryKey;column:id"`
	SKUCode             string         `gorm:"column:sku_code;size:64;not null"`
	CardTitle           string         `gorm:"column:card_title;size:128;not null"`
	CardSubtitle        string         `gorm:"column:card_subtitle;size:512;not null"`
	CardFeatures        datatypes.JSON `gorm:"column:card_features;type:json;not null"`
	ShareSeats          int            `gorm:"column:share_seats;not null;default:1"`
	ProductsCategoryID  *uint          `gorm:"column:products_category_id"`
	ProductsCategoryName string        `gorm:"column:products_category_name;size:32;not null"`
	SKUProductName      string         `gorm:"column:sku_product_name;size:128;not null"`
	LimitTokens         int64          `gorm:"column:limit_tokens;not null"`
	RPMLimit            int            `gorm:"column:rpm_limit;not null;default:0"`
	TPMLimit            *int           `gorm:"column:tpm_limit"`
	AllowedModels       datatypes.JSON `gorm:"column:allowed_models;type:json;not null"`
	BillingPeriod       string         `gorm:"column:billing_period;size:16;not null;default:month"`
	PriceCents          int64          `gorm:"column:price_cents;not null"`
	Currency            string         `gorm:"column:currency;size:3;not null;default:CNY"`
	IsHot               int            `gorm:"column:is_hot;not null;default:0"`
	HotTagName          string         `gorm:"column:hot_tag_name;size:32"`
	SortOrder           int            `gorm:"column:sort_order;not null;default:0"`
	Status              string         `gorm:"column:status;size:16;not null;default:on_sale"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
}

func (Product) TableName() string { return "products" }

type Payment struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	Code           string         `gorm:"column:code;size:32;not null"`
	DisplayName    string         `gorm:"column:display_name;size:64;not null"`
	Driver         string         `gorm:"column:driver;size:32;not null;default:epay"`
	APIBaseURL     string         `gorm:"column:api_base_url;size:512;not null"`
	MerchantID     string         `gorm:"column:merchant_id;size:64;not null"`
	MerchantSecret string         `gorm:"column:merchant_secret;size:256;not null"`
	AppID          *string        `gorm:"column:app_id;size:64"`
	ExtraJSON      datatypes.JSON `gorm:"column:extra_json;type:json"`
	NotifyPath     *string        `gorm:"column:notify_path;size:128"`
	IsEnabled      int            `gorm:"column:is_enabled;not null;default:0"`
	SortOrder      int            `gorm:"column:sort_order;not null;default:0"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (Payment) TableName() string { return "payments" }

type UserOrder struct {
	ID                uint       `gorm:"primaryKey;column:id"`
	OrderNo           string     `gorm:"column:order_no;size:64;not null"`
	UserID            uint       `gorm:"column:user_id;not null"`
	ProductID         uint       `gorm:"column:product_id;not null"`
	Quantity          int        `gorm:"column:quantity;not null;default:1"`
	UnitPriceCents    int64      `gorm:"column:unit_price_cents;not null"`
	Status            string     `gorm:"column:status;size:32;not null"`
	TotalAmountCents  int64      `gorm:"column:total_amount_cents;not null"`
	Currency          string     `gorm:"column:currency;size:3;not null;default:CNY"`
	EnterpriseInvoice int            `gorm:"column:enterprise_invoice;not null;default:0"`
	PayChannel        string         `gorm:"column:pay_channel;size:32"`
	OutTradeNo        string         `gorm:"column:out_trade_no;size:64"`
	ThirdTradeNo      *string        `gorm:"column:third_trade_no;size:128"`
	RawRequestJSON    datatypes.JSON `gorm:"column:raw_request_json;type:json"`
	RawNotifyJSON     datatypes.JSON `gorm:"column:raw_notify_json;type:json"`
	PaidAt            *time.Time     `gorm:"column:paid_at"`
	ExpireAt          time.Time  `gorm:"column:expire_at;not null"`
	ClosedAt          *time.Time `gorm:"column:closed_at"`
	FailReason        *string    `gorm:"column:fail_reason;size:512"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (UserOrder) TableName() string { return "user_orders" }

type PaymentCallback struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	Channel        string         `gorm:"column:channel;size:32;not null"`
	IdempotencyKey string         `gorm:"column:idempotency_key;size:128;not null"`
	PayloadJSON    datatypes.JSON `gorm:"column:payload_json;type:json;not null"`
	SignatureOK    int            `gorm:"column:signature_ok;not null;default:0"`
	ProcessResult  string         `gorm:"column:process_result;size:32;not null"`
	ProcessedAt    time.Time      `gorm:"column:processed_at"`
}

func (PaymentCallback) TableName() string { return "payment_callbacks" }

type UserWalletFlow struct {
	ID                uint      `gorm:"primaryKey;column:id"`
	UserID            uint      `gorm:"column:user_id;not null"`
	Type              string    `gorm:"column:type;size:32;not null"`
	AmountCents       int64     `gorm:"column:amount_cents;not null"`
	BalanceAfterCents *int64    `gorm:"column:balance_after_cents"`
	Currency          string    `gorm:"column:currency;size:3;not null;default:CNY"`
	RefType           *string   `gorm:"column:ref_type;size:32"`
	RefID             *uint     `gorm:"column:ref_id"`
	Remark            *string   `gorm:"column:remark;size:512"`
	CreatedAt         time.Time `gorm:"column:created_at"`
}

func (UserWalletFlow) TableName() string { return "user_wallet_flows" }

type UserRefund struct {
	ID          uint       `gorm:"primaryKey;column:id"`
	OrderID     uint       `gorm:"column:order_id;not null"`
	RefundNo    string     `gorm:"column:refund_no;size:64;not null"`
	AmountCents int64      `gorm:"column:amount_cents;not null"`
	Reason      *string    `gorm:"column:reason;size:512"`
	Status      string     `gorm:"column:status;size:32;not null"`
	RefundedAt  *time.Time `gorm:"column:refunded_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (UserRefund) TableName() string { return "user_refunds" }

type UserInvoice struct {
	ID          uint       `gorm:"primaryKey;column:id"`
	OrderID     uint       `gorm:"column:order_id;not null"`
	UserID      uint       `gorm:"column:user_id;not null"`
	InvoiceType string     `gorm:"column:invoice_type;size:32;not null"`
	Title       string     `gorm:"column:title;size:256;not null"`
	TaxNo       *string    `gorm:"column:tax_no;size:64"`
	AmountCents int64      `gorm:"column:amount_cents;not null"`
	Status      string     `gorm:"column:status;size:32;not null"`
	FileURL     *string    `gorm:"column:file_url;size:512"`
	IssuedAt    *time.Time `gorm:"column:issued_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (UserInvoice) TableName() string { return "user_invoices" }

type UserSubscription struct {
	ID               uint           `gorm:"primaryKey;column:id"`
	UserID           uint           `gorm:"column:user_id;not null"`
	ProductID        uint           `gorm:"column:product_id;not null"`
	Orders           datatypes.JSON `gorm:"column:orders;type:json;not null"`
	ProductsCategoryName string       `gorm:"column:products_category_name;size:32;not null"`
	SKUProductName       string       `gorm:"column:sku_product_name;size:128;not null"`
	BaseLimitTokens  int64          `gorm:"column:base_limit_tokens;not null"`
	LimitTokens      int64          `gorm:"column:limit_tokens;not null"`
	UsedTokens       int64          `gorm:"column:used_tokens;not null;default:0"`
	StartedAt        time.Time      `gorm:"column:started_at;not null"`
	ExpiresAt        time.Time      `gorm:"column:expires_at;not null"`
	PeriodStart      time.Time      `gorm:"column:period_start;not null"`
	PeriodEnd        time.Time      `gorm:"column:period_end;not null"`
	Status           string         `gorm:"column:status;size:32;not null;default:active"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
}

func (UserSubscription) TableName() string { return "user_subscriptions" }

type UserAPIKey struct {
	ID                 uint       `gorm:"primaryKey;column:id"`
	UserID             uint       `gorm:"column:user_id;not null"`
	UserSubscriptionID uint       `gorm:"column:user_subscription_id;not null"`
	KeyHash            string     `gorm:"column:key_hash;size:64;not null"`
	ProductsCategoryName string   `gorm:"column:products_category_name;size:32;not null"`
	LimitTokens          int64    `gorm:"column:limit_tokens;not null"`
	UsedTokens         int64      `gorm:"column:used_tokens;not null;default:0"`
	Status             string     `gorm:"column:status;size:32;not null;default:active"`
	RotatedAt          *time.Time `gorm:"column:rotated_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

func (UserAPIKey) TableName() string { return "user_api_keys" }

type UserNotification struct {
	ID                 uint       `gorm:"primaryKey;column:id"`
	UserID             uint       `gorm:"column:user_id;not null"`
	UserSubscriptionID *uint      `gorm:"column:user_subscription_id"`
	APIKeyID           *uint      `gorm:"column:api_key_id"`
	Channel            string     `gorm:"column:channel;size:32;not null"`
	TemplateCode       string     `gorm:"column:template_code;size:64;not null"`
	Status             string     `gorm:"column:status;size:32;not null"`
	SentAt             *time.Time `gorm:"column:sent_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
}

func (UserNotification) TableName() string { return "user_notifications" }

type UserAccessLog struct {
	ID                 uint      `gorm:"primaryKey;column:id"`
	RequestID          string    `gorm:"column:request_id;size:64;not null"`
	UserID             *uint     `gorm:"column:user_id"`
	UserSubscriptionID *uint     `gorm:"column:user_subscription_id"`
	APIKeyID           *uint     `gorm:"column:api_key_id"`
	ProductsCategoryName string  `gorm:"column:products_category_name;size:32;not null"`
	SKUProductName       *string `gorm:"column:sku_product_name;size:128"`
	UpstreamInfoID     *uint     `gorm:"column:upstream_info_id"`
	Model              *string   `gorm:"column:model;size:128"`
	HTTPMethod         string    `gorm:"column:http_method;size:16;not null"`
	Path               string    `gorm:"column:path;size:512;not null"`
	ClientIP           *string   `gorm:"column:client_ip;size:64"`
	GatewayStatus      int16     `gorm:"column:gateway_status;not null"`
	UpstreamStatus     *int16    `gorm:"column:upstream_status"`
	LatencyMs          int       `gorm:"column:latency_ms;not null;default:0"`
	TokensPrompt       *int      `gorm:"column:tokens_prompt"`
	TokensCompletion   *int      `gorm:"column:tokens_completion"`
	TokensTotal        *int      `gorm:"column:tokens_total"`
	IsStream           int       `gorm:"column:is_stream;not null;default:0"`
	ErrorCode          *string   `gorm:"column:error_code;size:32"`
	CreatedAt          time.Time `gorm:"column:created_at"`
}

func (UserAccessLog) TableName() string { return "user_access_logs" }

type TutorialCategory struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Code      string    `gorm:"column:code;size:32;not null"`
	Name      string    `gorm:"column:name;size:64;not null"`
	Sort      int       `gorm:"column:sort;not null;default:0"`
	Status    string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (TutorialCategory) TableName() string { return "tutorial_category" }

type TutorialArticle struct {
	ID          uint           `gorm:"primaryKey;column:id"`
	CategoryID  uint           `gorm:"column:category_id;not null"`
	Slug        string         `gorm:"column:slug;size:128;not null"`
	Title       string         `gorm:"column:title;size:256;not null"`
	Excerpt     string         `gorm:"column:excerpt;size:512;not null"`
	Body        datatypes.JSON `gorm:"column:body;type:json;not null"`
	PublishedAt time.Time      `gorm:"column:published_at;type:date"`
	Sort        int            `gorm:"column:sort;not null;default:0"`
	Status      string         `gorm:"column:status;size:16;not null;default:published"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (TutorialArticle) TableName() string { return "tutorial_article" }

type EnterpriseInquiry struct {
	ID          uint      `gorm:"primaryKey;column:id"`
	CompanyName string    `gorm:"column:company_name;size:256;not null"`
	ContactName string    `gorm:"column:contact_name;size:128;not null"`
	Phone       string    `gorm:"column:phone;size:32;not null"`
	Email       *string   `gorm:"column:email;size:255"`
	Status      string    `gorm:"column:status;size:32;not null;default:pending"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (EnterpriseInquiry) TableName() string { return "enterprise_inquiry" }

type EnterpriseProduct struct {
	ID          uint           `gorm:"primaryKey;column:id"`
	Code        string         `gorm:"column:code;size:32;not null"`
	Name        string         `gorm:"column:name;size:64;not null"`
	Badge       *string        `gorm:"column:badge;size:32"`
	PriceHint   string         `gorm:"column:price_hint;size:128;not null"`
	Seats       string         `gorm:"column:seats;size:128;not null"`
	Features    datatypes.JSON `gorm:"column:features;type:json;not null"`
	Tagline     string         `gorm:"column:tagline;size:256;not null"`
	ButtonLabel string         `gorm:"column:button_label;size:64;not null;default:获取报价"`
	IsFeatured  int            `gorm:"column:is_featured;not null;default:0"`
	Sort        int            `gorm:"column:sort;not null;default:0"`
	Status      string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (EnterpriseProduct) TableName() string { return "enterprise_products" }
