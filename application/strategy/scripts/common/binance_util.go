package common

import (
	"fmt"
	"strings"

	appcommon "github.com/gph-tech/fgmm-strategy-bitfinex/application/common"
	httpentity "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/shopspring/decimal"
)

// ResolveSymbol 任务上的 Binance 交易对。
func ResolveSymbol(task entity.LiquidationTask) string {
	if task.Symbol != nil && *task.Symbol != "" {
		return strings.ToUpper(*task.Symbol)
	}
	return ""
}

// CurrentTier 任务当前档位，缺省 P2。
func CurrentTier(task entity.LiquidationTask) string {
	if task.CurrentTier != nil && *task.CurrentTier != "" {
		return *task.CurrentTier
	}
	return appcommon.LiquidationTierP2
}

// truncateQty 数量截断到 8 位小数。
func truncateQty(q decimal.Decimal) decimal.Decimal {
	return q.Truncate(8)
}

// FormatDecimal 下单用 8 位小数字符串。
func FormatDecimal(d decimal.Decimal) string {
	return d.StringFixed(8)
}

// MapBinanceOrderStatus Binance 状态映射为内部子单状态。
func MapBinanceOrderStatus(st string) string {
	switch st {
	case httpentity.BinanceOrderStatusNew:
		return entity.ExchangeOrderStatusNew
	case httpentity.BinanceOrderStatusPartiallyFilled:
		return entity.ExchangeOrderStatusPartiallyFilled
	case httpentity.BinanceOrderStatusFilled:
		return entity.ExchangeOrderStatusFilled
	case httpentity.BinanceOrderStatusCanceled, httpentity.BinanceOrderStatusExpired:
		return entity.ExchangeOrderStatusCancelled
	case httpentity.BinanceOrderStatusRejected:
		return entity.ExchangeOrderStatusRejected
	default:
		return entity.ExchangeOrderStatusUnknown
	}
}

// NewClientOrderID 生成 Binance clientOrderId（最长 36）。
func NewClientOrderID(liquidationID string, seq uint) string {
	id := fmt.Sprintf("liq-%s-%d", liquidationID, seq)
	if len(id) > 36 {
		id = id[:36]
	}
	return id
}
