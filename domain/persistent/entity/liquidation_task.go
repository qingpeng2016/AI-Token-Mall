package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	LiquidationStatusLiquidating = "liquidating" // 清算执行中
	LiquidationStatusPaused      = "paused"      // 已暂停
	LiquidationStatusSettled     = "settled"     // 已清偿（债务已覆盖）
	LiquidationStatusBadDebt     = "bad_debt"    // 坏账（抵押卖尽仍欠债）
	LiquidationStatusCancelled   = "cancelled"   // 已取消
	LiquidationStatusFailed      = "failed"      // 失败终态
)

const (
	CallbackStatusPending = "pending" // 待向 BeTrust 回调
	CallbackStatusSuccess = "success" // 回调已成功
	CallbackStatusFailed  = "failed"  // 回调失败（可重试）
)

const (
	InFlightOrdersClearedNo  = 0 // 在途挂单尚未确认清理完成
	InFlightOrdersClearedYes = 1 // 在途挂单已清理（或无在途）
)

type LiquidationTask struct {
	ID                         uint             `gorm:"primaryKey;column:id"`
	LiquidationID              string           `gorm:"column:liquidation_id"`
	LoanOrderID                string           `gorm:"column:loan_order_id"`
	UID                        string           `gorm:"column:uid"`
	Status                     string           `gorm:"column:status"`
	EndReason                  *string          `gorm:"column:end_reason"`
	OriginalRatio              decimal.Decimal  `gorm:"column:original_ratio"`
	TriggerRatio               decimal.Decimal  `gorm:"column:trigger_ratio"`
	TriggerMarkPrice           decimal.Decimal  `gorm:"column:trigger_mark_price"`
	TriggerCollateralAsset     string           `gorm:"column:trigger_collateral_asset"`
	TriggerCollateralQty       decimal.Decimal  `gorm:"column:trigger_collateral_qty"`
	TriggerCollateralDueAmount decimal.Decimal  `gorm:"column:trigger_collateral_due_amount"`
	Symbol                     *string          `gorm:"column:symbol"`
	CurrentTier                *string          `gorm:"column:current_tier"`
	CurrentTierReasonJSON      *string          `gorm:"column:current_tier_reason_json"`
	FeeRate                    decimal.Decimal  `gorm:"column:fee_rate"`
	ExchangeCode               *string          `gorm:"column:exchange_code"`
	LastMidPrice               *decimal.Decimal `gorm:"column:last_mid_price"`
	LastMidPriceAt             *time.Time       `gorm:"column:last_mid_price_at"`
	OrderFailStreak            uint             `gorm:"column:order_fail_streak"`
	NextQuoteAt                *time.Time       `gorm:"column:next_quote_at"`
	CallbackURL                string           `gorm:"column:callback_url"`
	CallbackStatus             *string          `gorm:"column:callback_status"`
	CallbackAt                 *time.Time       `gorm:"column:callback_at"`
	CancelReason               *string          `gorm:"column:cancel_reason"`
	CancelRequestedAt          *time.Time       `gorm:"column:cancel_requested_at"`
	CancelResult               *string          `gorm:"column:cancel_result"`
	InFlightOrdersCleared      int              `gorm:"column:in_flight_orders_cleared;not null;default:0"` // 在途挂单是否已清理 0否 1是
	StartedAt                  *time.Time       `gorm:"column:started_at"`
	CreatedAt                  time.Time        `gorm:"column:created_at"`
	UpdatedAt                  time.Time        `gorm:"column:updated_at"`
}

func (LiquidationTask) TableName() string {
	return "liquidation_task"
}
