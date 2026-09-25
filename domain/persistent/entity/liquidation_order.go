package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	ExchangeOrderStatusNew             = "new"              // 已落库，待报送或待回执
	ExchangeOrderStatusUnknown         = "unknown"          // 交易所状态未知
	ExchangeOrderStatusPartiallyFilled = "partially_filled" // 部分成交
	ExchangeOrderStatusFilled          = "filled"           // 全部成交
	ExchangeOrderStatusCancelled       = "cancelled"        // 已撤销
	ExchangeOrderStatusRejected        = "rejected"         // 拒单
)

type LiquidationOrder struct {
	ID              uint             `gorm:"primaryKey;column:id"`
	LiquidationID   string           `gorm:"column:liquidation_id"`
	BatchNo         string           `gorm:"column:batch_no"`
	ClientOrderID   string           `gorm:"column:client_order_id"`
	ExchangeCode    string           `gorm:"column:exchange_code"`
	ExchangeOrderID *string          `gorm:"column:exchange_order_id"`
	Symbol          string           `gorm:"column:symbol"`
	Side            string           `gorm:"column:side"`
	OrderType       string           `gorm:"column:order_type"`
	Tier            *string          `gorm:"column:tier"`
	FloorPrice      *decimal.Decimal `gorm:"column:floor_price"`
	Price           decimal.Decimal  `gorm:"column:price"`
	Quantity        decimal.Decimal  `gorm:"column:quantity"`
	FilledQty       decimal.Decimal  `gorm:"column:filled_qty"`
	AvgPrice        *decimal.Decimal `gorm:"column:avg_price"`
	Status          string           `gorm:"column:status"`
	FailureCode     *string          `gorm:"column:failure_code"`
	FailureReason   *string          `gorm:"column:failure_reason"`
	SubmitRaw       *string          `gorm:"column:submit_raw;type:json"`
	ExecSnapshot    *string          `gorm:"column:exec_snapshot;type:json"`
	LastReportRaw   *string          `gorm:"column:last_report_raw;type:json"`
	SubmittedAt     *time.Time       `gorm:"column:submitted_at"`
	FinishedAt      *time.Time       `gorm:"column:finished_at"`
	CreatedAt       time.Time        `gorm:"column:created_at"`
	UpdatedAt       time.Time        `gorm:"column:updated_at"`
}

const ExchangeCodeBinance = "binance" // 币安现货

func (LiquidationOrder) TableName() string {
	return "liquidation_orders"
}
