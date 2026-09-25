package entity

import "time"

type LiquidationTaskEvent struct {
	ID            uint      `gorm:"primaryKey;column:id"`
	LiquidationID string    `gorm:"column:liquidation_id"`
	EventType     string    `gorm:"column:event_type"`
	FromStatus    *string   `gorm:"column:from_status"`
	ToStatus      *string   `gorm:"column:to_status"`
	Message       *string   `gorm:"column:message"`
	Extra         *string   `gorm:"column:extra;type:json"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (LiquidationTaskEvent) TableName() string {
	return "liquidation_task_event"
}

const (
	TaskEventSignalReceived = "signal_received" // 收到 BeTrust 触发/创建
	TaskEventTierChanged    = "tier_changed"    // 执行档位变更
	TaskEventOrderSubmit    = "order_submit"    // 向交易所提交卖单
	TaskEventFill           = "fill"            // 成交更新
	TaskEventPause          = "pause"           // 任务暂停
	TaskEventCancel         = "cancel"          // 任务取消
	TaskEventSettled        = "settled"         // 进入已清偿终态
	TaskEventBadDebt        = "bad_debt"        // 进入坏账终态
	TaskEventAlert          = "alert"           // 告警事件
)
