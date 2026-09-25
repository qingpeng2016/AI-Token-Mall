package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type LiquidationTrader struct {
	ID                uint            `gorm:"primaryKey;column:id"`
	LiquidationID     string          `gorm:"column:liquidation_id"`
	BatchNo           string          `gorm:"column:batch_no"`
	LiquidationOrderID uint           `gorm:"column:liquidation_order_id"`
	ExchangeCode      string          `gorm:"column:exchange_code"`
	ExchangeTradeID   string          `gorm:"column:exchange_trade_id"`
	Symbol            string          `gorm:"column:symbol"`
	Price             decimal.Decimal `gorm:"column:price"`
	Quantity          decimal.Decimal `gorm:"column:quantity"`
	Amount            decimal.Decimal `gorm:"column:amount"`
	Fee               decimal.Decimal `gorm:"column:fee"`
	FeeAsset          *string         `gorm:"column:fee_asset"`
	IsMaker           *bool           `gorm:"column:is_maker"`
	TradeTime         time.Time       `gorm:"column:trade_time"`
	RawPayload        string          `gorm:"column:raw_payload;type:json"`
	CreatedAt         time.Time       `gorm:"column:created_at"`
}

func (LiquidationTrader) TableName() string {
	return "liquidation_traders"
}
