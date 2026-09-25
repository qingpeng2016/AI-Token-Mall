package common

import "strings"

// 抵押币种（与 BeTrust collateral_asset、PRD 口径一致，统一大写）
const (
	CollateralAssetBTC = "BTC" // 比特币抵押
	CollateralAssetETH = "ETH" // 以太坊抵押
)

// 强平执行档位
const (
	LiquidationTierP0 = "P0" // 最紧急（优先执行）
	LiquidationTierP1 = "P1" // 次紧急
	LiquidationTierP2 = "P2" // 常规
)

// TierEmergencyRank 数值越小越紧急，用于执行排序。
func TierEmergencyRank(tier string) uint {
	switch strings.ToUpper(strings.TrimSpace(tier)) {
	case LiquidationTierP0:
		return 0
	case LiquidationTierP1:
		return 1
	default:
		return 2
	}
}

// NormalizeCollateralAsset 用于比较/分支（trim + 大写）
func NormalizeCollateralAsset(asset string) string {
	return strings.ToUpper(strings.TrimSpace(asset))
}
