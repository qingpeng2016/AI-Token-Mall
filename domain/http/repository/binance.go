package repository

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
)

// BinanceRepo 强平脚本所需的 Binance Spot REST（公共读 + 现货 signed 交易/查单）
type BinanceRepo interface {
	Ping(ctx context.Context, auth *entity.BinanceAuth) error
	GetBookTicker(ctx context.Context, auth *entity.BinanceAuth, symbol string) (*entity.BinanceBookTicker, error)
	CreateSpotOrder(ctx context.Context, auth entity.BinanceAuth, req entity.BinanceCreateOrderReq) (*entity.BinanceOrder, error)
	GetSpotOrder(ctx context.Context, auth entity.BinanceAuth, q entity.BinanceGetOrderQuery) (*entity.BinanceOrder, error)
	GetMyTrades(ctx context.Context, auth entity.BinanceAuth, q entity.BinanceMyTradesQuery) ([]entity.BinanceTrade, error)
	GetDepth(ctx context.Context, auth *entity.BinanceAuth, symbol string, limit int) (*entity.BinanceDepth, error)
	GetExchangeInfo(ctx context.Context, auth *entity.BinanceAuth, symbol string) (*entity.BinanceSymbolInfo, error)
	CancelSpotOrder(ctx context.Context, auth entity.BinanceAuth, symbol, origClientOrderID string) (*entity.BinanceOrder, error)
}
