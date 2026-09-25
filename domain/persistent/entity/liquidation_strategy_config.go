package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type LiquidationStrategyConfig struct {
	ID                          uint            `gorm:"primaryKey;column:id"`
	GlobalDepthCapRatio         decimal.Decimal `gorm:"column:global_depth_cap_ratio"`
	OrderDepthRatio             decimal.Decimal `gorm:"column:order_depth_ratio"`
	MaxConcurrentTasksPerSymbol uint            `gorm:"column:max_concurrent_tasks_per_symbol"`
	MaxNotionalUSDPerSymbol     decimal.Decimal `gorm:"column:max_notional_usd_per_symbol"`
	MaxSingleOrderDurationSec   uint            `gorm:"column:max_single_order_duration_sec"`
	MaxOrdersPerTask            uint            `gorm:"column:max_orders_per_task"`
	ConsecutiveOrderFailPause   uint            `gorm:"column:consecutive_order_fail_pause"`
	StaleQuoteSec               uint            `gorm:"column:stale_quote_sec"`
	HardBoundaryBTCPct          decimal.Decimal `gorm:"column:hard_boundary_btc_pct"`
	HardBoundaryETHPct          decimal.Decimal `gorm:"column:hard_boundary_eth_pct"`
	HardBoundaryDefaultPct      decimal.Decimal `gorm:"column:hard_boundary_default_pct"`
	IsEnabled                   int             `gorm:"column:is_enabled"`
	UpdatedAt                   time.Time       `gorm:"column:updated_at"`
}

func (LiquidationStrategyConfig) TableName() string {
	return "liquidation_config_strategy"
}

type LiquidationTierConfig struct {
	Tier                       string           `gorm:"primaryKey;column:tier"`
	TierRank                   uint             `gorm:"column:tier_rank"`
	SafetyBufferLtPct          *decimal.Decimal `gorm:"column:safety_buffer_lt_pct"`
	SafetyBufferGtePct         *decimal.Decimal `gorm:"column:safety_buffer_gte_pct"`
	SafetyBufferBetweenLowPct  *decimal.Decimal `gorm:"column:safety_buffer_between_low_pct"`
	SafetyBufferBetweenHighPct *decimal.Decimal `gorm:"column:safety_buffer_between_high_pct"`
	ExecTimeExceedSec          *uint            `gorm:"column:exec_time_exceed_sec"`
	ExecTimeUnderSec           *uint            `gorm:"column:exec_time_under_sec"`
	ConsecutiveIOCUnfilled     *uint            `gorm:"column:consecutive_ioc_unfilled"`
	PriceDropWindowSec         *uint            `gorm:"column:price_drop_window_sec"`
	PriceDropPct               *decimal.Decimal `gorm:"column:price_drop_pct"`
	BidOffsetBTCPct            decimal.Decimal  `gorm:"column:bid_offset_btc_pct"`
	BidOffsetETHPct            decimal.Decimal  `gorm:"column:bid_offset_eth_pct"`
	BidOffsetDefaultPct        decimal.Decimal  `gorm:"column:bid_offset_default_pct"`
	QuoteIntervalSec           float64          `gorm:"column:quote_interval_sec"`
	IsEnabled                  int              `gorm:"column:is_enabled"`
}

func (LiquidationTierConfig) TableName() string {
	return "liquidation_config_tier"
}
