package entity

import "time"

const (
	CallbackLogStatusPending = "pending" // 单次回调尝试进行中
	CallbackLogStatusSuccess = "success" // 单次回调 HTTP 成功
	CallbackLogStatusFailed  = "failed"  // 单次回调失败
)

type LiquidationCallbackLog struct {
	ID            uint       `gorm:"primaryKey;column:id"`
	LiquidationID string     `gorm:"column:liquidation_id"`
	CallbackURL   string     `gorm:"column:callback_url"`
	Payload       string     `gorm:"column:payload;type:json"`
	HTTPStatus    *int       `gorm:"column:http_status"`
	ResponseBody  *string    `gorm:"column:response_body"`
	Status        string     `gorm:"column:status"`
	RetryCount    uint       `gorm:"column:retry_count"`
	NextRetryAt   *time.Time `gorm:"column:next_retry_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (LiquidationCallbackLog) TableName() string {
	return "liquidation_log_callback"
}
