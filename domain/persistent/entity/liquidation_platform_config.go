package entity

import "time"

// LiquidationPlatformConfig 对接平台（如 BeTrust）鉴权与回调配置。
type LiquidationPlatformConfig struct {
	ID           uint      `gorm:"primaryKey;column:id;type:bigint unsigned;not null;autoIncrement"`
	PlatformName string    `gorm:"column:platform_name;type:varchar(100);not null;comment:对接平台名称"`
	SignKey      string    `gorm:"column:sign_key;type:varchar(255);not null;comment:与 BeTrust 共用的 HMAC sign 密钥"`
	BnAPIKey     string    `gorm:"column:bn_api_key;type:varchar(255);not null;default:'';comment:Binance 现货 API Key"`
	BnAPISecret  string    `gorm:"column:bn_api_secret;type:varchar(512);not null;default:'';comment:Binance 现货 API Secret"`
	CallbackURL  string    `gorm:"column:callback_url;type:varchar(512);not null;comment:清算终态 POST 回调地址"`
	IsEnabled    int       `gorm:"column:is_enabled;type:tinyint(1);not null;default:1"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (LiquidationPlatformConfig) TableName() string {
	return "liquidation_platform_config"
}
