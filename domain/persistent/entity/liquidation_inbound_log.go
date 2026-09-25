package entity

import "time"

const (
	InboundAPIActionTrigger = "trigger" // 创建/触发清算任务
	InboundAPIActionCancel  = "cancel"  // 取消清算任务
)

type LiquidationInboundLog struct {
	ID             uint      `gorm:"primaryKey;column:id"`
	LiquidationID  string    `gorm:"column:liquidation_id"`
	LoanOrderID    string    `gorm:"column:loan_order_id"`
	APIPath        string    `gorm:"column:api_path"`
	APIAction      string    `gorm:"column:api_action"`
	EventTimeMs    int64     `gorm:"column:event_time_ms"`
	IdempotencyKey string    `gorm:"column:idempotency_key"`
	RequestBody    string    `gorm:"column:request_body;type:json"`
	ResponseBody   *string   `gorm:"column:response_body;type:json"`
	HTTPStatus     *int      `gorm:"column:http_status"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (LiquidationInboundLog) TableName() string {
	return "liquidation_log_inbound"
}
