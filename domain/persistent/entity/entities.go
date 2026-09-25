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
	UpstreamName         string     `gorm:"column:upstream_name;size:32;not null"`
	UpstreamProduct      string     `gorm:"column:upstream_product;size:128;not null"`
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

type Product struct {
	ID                  uint           `gorm:"primaryKey;column:id"`
	SKUCode             string         `gorm:"column:sku_code;size:64;not null"`
	MarketingTier       string         `gorm:"column:marketing_tier;size:128;not null"`
	UpstreamName        string         `gorm:"column:upstream_name;size:32;not null"`
	UpstreamProduct     string         `gorm:"column:upstream_product;size:128;not null"`
	LimitTokens         int64          `gorm:"column:limit_tokens;not null"`
	RPMLimit            int            `gorm:"column:rpm_limit;not null;default:0"`
	TPMLimit            *int           `gorm:"column:tpm_limit"`
	AllowedModels       datatypes.JSON `gorm:"column:allowed_models;type:json;not null"`
	ProductType         string         `gorm:"column:product_type;size:32;not null;default:subscription"`
	BillingPeriod       string         `gorm:"column:billing_period;size:16;not null;default:month"`
	PriceCents          int64          `gorm:"column:price_cents;not null"`
	Currency            string         `gorm:"column:currency;size:3;not null;default:CNY"`
	CompareAtPriceCents *int64         `gorm:"column:compare_at_price_cents"`
	HighlightsJSON      datatypes.JSON `gorm:"column:highlights_json;type:json"`
	IsHot               int            `gorm:"column:is_hot;not null;default:0"`
	IsAPIEnabled        int            `gorm:"column:is_api_enabled;not null;default:1"`
	TopupTokenAmount    *int64         `gorm:"column:topup_token_amount"`
	SortOrder           int            `gorm:"column:sort_order;not null;default:0"`
	Status              string         `gorm:"column:status;size:16;not null;default:on_sale"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
}

func (Product) TableName() string { return "products" }

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
	EnterpriseInvoice int        `gorm:"column:enterprise_invoice;not null;default:0"`
	PaidAt            *time.Time `gorm:"column:paid_at"`
	ExpireAt          time.Time  `gorm:"column:expire_at;not null"`
	ClosedAt          *time.Time `gorm:"column:closed_at"`
	FailReason        *string    `gorm:"column:fail_reason;size:512"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (UserOrder) TableName() string { return "user_orders" }

type UserPayment struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	OrderID        uint           `gorm:"column:order_id;not null"`
	Channel        string         `gorm:"column:channel;size:32;not null"`
	OutTradeNo     string         `gorm:"column:out_trade_no;size:64;not null"`
	ThirdTradeNo   *string        `gorm:"column:third_trade_no;size:128"`
	AmountCents    int64          `gorm:"column:amount_cents;not null"`
	Status         string         `gorm:"column:status;size:32;not null"`
	PaidAt         *time.Time     `gorm:"column:paid_at"`
	RawRequestJSON datatypes.JSON `gorm:"column:raw_request_json;type:json"`
	RawNotifyJSON  datatypes.JSON `gorm:"column:raw_notify_json;type:json"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (UserPayment) TableName() string { return "user_payments" }

type PaymentCallback struct {
	ID            uint           `gorm:"primaryKey;column:id"`
	Channel       string         `gorm:"column:channel;size:32;not null"`
	IdempotencyKey string        `gorm:"column:idempotency_key;size:128;not null"`
	PayloadJSON   datatypes.JSON `gorm:"column:payload_json;type:json;not null"`
	SignatureOK   int            `gorm:"column:signature_ok;not null;default:0"`
	ProcessResult string         `gorm:"column:process_result;size:32;not null"`
	ProcessedAt   time.Time      `gorm:"column:processed_at"`
}

func (PaymentCallback) TableName() string { return "payment_callbacks" }

type UserRefund struct {
	ID          uint       `gorm:"primaryKey;column:id"`
	OrderID     uint       `gorm:"column:order_id;not null"`
	PaymentID   *uint      `gorm:"column:payment_id"`
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
	UpstreamName     string         `gorm:"column:upstream_name;size:32;not null"`
	UpstreamProduct  string         `gorm:"column:upstream_product;size:128;not null"`
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
	UpstreamName       string     `gorm:"column:upstream_name;size:32;not null"`
	LimitTokens        int64      `gorm:"column:limit_tokens;not null"`
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
	UpstreamName       string    `gorm:"column:upstream_name;size:32;not null"`
	UpstreamProduct    *string   `gorm:"column:upstream_product;size:128"`
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
