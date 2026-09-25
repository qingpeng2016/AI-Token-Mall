package common

import (
	"context"
	"time"

	appcommon "github.com/gph-tech/fgmm-strategy-bitfinex/application/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/shopspring/decimal"
)

// TierContext 档位判定输入（安全垫、耗时、IOC 连空、跌幅）。
type TierContext struct {
	SafetyBuffer       decimal.Decimal
	ExecSec            uint
	ConsecutiveIOCZero int
	PriceDropPct       decimal.Decimal
}

// LoadTierConfigs 读 DB 档位；空则内置默认三档。
func LoadTierConfigs(ctx context.Context, repo repository.LiquidationExecutionConfigRepo) []entity.LiquidationTierConfig {
	rows, err := repo.ListTiers(ctx)
	if err != nil || len(rows) == 0 {
		return defaultTierRows()
	}
	return rows
}

// defaultTierRows PRD 默认 P0/P1/P2 参数。
func defaultTierRows() []entity.LiquidationTierConfig {
	lt8 := decimal.RequireFromString("0.08")
	gte15 := decimal.RequireFromString("0.15")
	drop05 := decimal.RequireFromString("0.005")
	ex30 := uint(30)
	ex15 := uint(15)
	exUnder15 := uint(15)
	ioc3 := uint(3)
	win5 := uint(5)
	return []entity.LiquidationTierConfig{
		{Tier: appcommon.LiquidationTierP0, TierRank: 0, SafetyBufferLtPct: &lt8, ExecTimeExceedSec: &ex30,
			BidOffsetBTCPct: decimal.RequireFromString("0.015"), BidOffsetETHPct: decimal.RequireFromString("0.02"), BidOffsetDefaultPct: decimal.RequireFromString("0.015"), QuoteIntervalSec: 0.5, IsEnabled: 1},
		{Tier: appcommon.LiquidationTierP1, TierRank: 1, SafetyBufferBetweenLowPct: &lt8, SafetyBufferBetweenHighPct: &gte15,
			ExecTimeExceedSec: &ex15, ConsecutiveIOCUnfilled: &ioc3, PriceDropWindowSec: &win5, PriceDropPct: &drop05,
			BidOffsetBTCPct: decimal.RequireFromString("0.008"), BidOffsetETHPct: decimal.RequireFromString("0.008"), BidOffsetDefaultPct: decimal.RequireFromString("0.008"), QuoteIntervalSec: 1, IsEnabled: 1},
		{Tier: appcommon.LiquidationTierP2, TierRank: 2, SafetyBufferGtePct: &gte15, ExecTimeUnderSec: &exUnder15,
			BidOffsetBTCPct: decimal.RequireFromString("0.003"), BidOffsetETHPct: decimal.RequireFromString("0.003"), BidOffsetDefaultPct: decimal.RequireFromString("0.003"), QuoteIntervalSec: 1, IsEnabled: 1},
	}
}

// ComputeSafetyBuffer (spot - 破产价) / spot。
func ComputeSafetyBuffer(spot, totalDue, collateral decimal.Decimal) decimal.Decimal {
	if !spot.IsPositive() || !collateral.IsPositive() {
		return decimal.Zero
	}
	bankruptcy := totalDue.Div(collateral)
	return spot.Sub(bankruptcy).Div(spot)
}

// ExecSeconds 任务已执行秒数。
func ExecSeconds(task entity.LiquidationTask) uint {
	if task.StartedAt == nil {
		return 0
	}
	sec := time.Since(*task.StartedAt).Seconds()
	if sec < 0 {
		return 0
	}
	return uint(sec)
}

// PriceDropPct 相对上次 mid 的跌幅比例。
func PriceDropPct(task entity.LiquidationTask, currentMid decimal.Decimal) decimal.Decimal {
	if task.LastMidPrice == nil || !task.LastMidPrice.IsPositive() || !currentMid.IsPositive() {
		return decimal.Zero
	}
	prev := *task.LastMidPrice
	return prev.Sub(currentMid).Div(prev)
}

// tierMatches 单条档位规则是否命中。
func tierMatches(cfg entity.LiquidationTierConfig, c TierContext) bool {
	if cfg.SafetyBufferLtPct != nil && c.SafetyBuffer.LessThan(*cfg.SafetyBufferLtPct) {
		return true
	}
	if cfg.SafetyBufferBetweenLowPct != nil && cfg.SafetyBufferBetweenHighPct != nil {
		if !c.SafetyBuffer.LessThan(*cfg.SafetyBufferBetweenLowPct) && c.SafetyBuffer.LessThan(*cfg.SafetyBufferBetweenHighPct) {
			return true
		}
	}
	if cfg.SafetyBufferGtePct != nil && cfg.ExecTimeUnderSec != nil {
		if c.SafetyBuffer.GreaterThanOrEqual(*cfg.SafetyBufferGtePct) && c.ExecSec < *cfg.ExecTimeUnderSec {
			return true
		}
	}
	if cfg.ExecTimeExceedSec != nil && c.ExecSec > *cfg.ExecTimeExceedSec {
		return true
	}
	if cfg.ConsecutiveIOCUnfilled != nil && c.ConsecutiveIOCZero >= int(*cfg.ConsecutiveIOCUnfilled) {
		return true
	}
	if cfg.PriceDropPct != nil && c.PriceDropPct.GreaterThan(*cfg.PriceDropPct) {
		return true
	}
	return false
}

// ResolveTierForTask 按任务状态与 mid 重算档位（只升不降）。
func ResolveTierForTask(task entity.LiquidationTask, tiers []entity.LiquidationTierConfig, mid decimal.Decimal, consecutiveIOCZero int) string {
	tier, _ := ResolveTierForTaskWithExplain(task, tiers, mid, consecutiveIOCZero)
	return tier
}

// ResolveTier 只升不降：命中档位中取 tier_rank 最小（最紧急）。
func ResolveTier(current string, tiers []entity.LiquidationTierConfig, c TierContext) string {
	bestRank := uint(99)
	bestTier := appcommon.LiquidationTierP2
	if current != "" {
		for _, t := range tiers {
			if t.Tier == current {
				bestRank = t.TierRank
				bestTier = current
				break
			}
		}
	}
	for _, t := range tiers {
		if t.IsEnabled != 1 {
			continue
		}
		if tierMatches(t, c) && t.TierRank < bestRank {
			bestRank = t.TierRank
			bestTier = t.Tier
		}
	}
	return bestTier
}

// BidOffsetForTier 按抵押币种取档位买一偏移。
func BidOffsetForTier(cfg entity.LiquidationTierConfig, collateralAsset string) decimal.Decimal {
	switch appcommon.NormalizeCollateralAsset(collateralAsset) {
	case appcommon.CollateralAssetBTC:
		return cfg.BidOffsetBTCPct
	case appcommon.CollateralAssetETH:
		return cfg.BidOffsetETHPct
	default:
		return cfg.BidOffsetDefaultPct
	}
}

// TierConfigByName 按档位名查配置行。
func TierConfigByName(tiers []entity.LiquidationTierConfig, name string) *entity.LiquidationTierConfig {
	for i := range tiers {
		if tiers[i].Tier == name {
			return &tiers[i]
		}
	}
	return nil
}
