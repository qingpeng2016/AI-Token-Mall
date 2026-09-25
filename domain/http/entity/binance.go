// 官方参考：https://developers.binance.com/docs/binance-spot-api-docs/rest-api
package entity

import "encoding/json"

const (
	BinanceSideSell = "SELL" // 卖（清算侧固定用此方向）
	BinanceSideBuy  = "BUY"  // 买

	BinanceOrderTypeLimit = "LIMIT" // 限价单
	BinanceTIFIOC         = "IOC"   // 立即成交否则取消

	BinanceOrderStatusNew             = "NEW"              // 已接受
	BinanceOrderStatusPartiallyFilled = "PARTIALLY_FILLED" // 部分成交
	BinanceOrderStatusFilled          = "FILLED"           // 全部成交
	BinanceOrderStatusCanceled        = "CANCELED"         // 已撤销
	BinanceOrderStatusRejected        = "REJECTED"         // 拒单
	BinanceOrderStatusExpired         = "EXPIRED"          // 过期（如 IOC 未成交部分）
)

// BinanceAuth 现货 signed 接口密钥；BaseURL 空则用配置默认（如 https://api.binance.com）
type BinanceAuth struct {
	APIKey    string
	APISecret string
	BaseURL   string
}

// BinanceError 业务错误 JSON
type BinanceError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *BinanceError) Error() string {
	if e == nil {
		return "binance error"
	}
	return e.Msg
}

// BinanceDepth GET /api/v3/depth
type BinanceDepth struct {
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
}

// BinanceSymbolFilter exchangeInfo filters subset
type BinanceSymbolFilter struct {
	FilterType  string `json:"filterType"`
	MinPrice    string `json:"minPrice,omitempty"`
	MaxPrice    string `json:"maxPrice,omitempty"`
	TickSize    string `json:"tickSize,omitempty"`
	MinQty      string `json:"minQty,omitempty"`
	MaxQty      string `json:"maxQty,omitempty"`
	StepSize    string `json:"stepSize,omitempty"`
	MinNotional string `json:"minNotional,omitempty"`
}

type BinanceSymbolInfo struct {
	Symbol  string                `json:"symbol"`
	Status  string                `json:"status"`
	Filters []BinanceSymbolFilter `json:"filters"`
}

type BinanceExchangeInfo struct {
	Symbols []BinanceSymbolInfo `json:"symbols"`
}

// BinanceBookTicker GET /api/v3/ticker/bookTicker
type BinanceBookTicker struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	BidQty   string `json:"bidQty"`
	AskPrice string `json:"askPrice"`
	AskQty   string `json:"askQty"`
}

// BinanceCreateOrderReq POST /api/v3/order
type BinanceCreateOrderReq struct {
	Symbol           string
	Side             string
	Type             string
	TimeInForce      string
	Quantity         string
	Price            string
	NewClientOrderID string
}

// BinanceGetOrderQuery GET /api/v3/order
type BinanceGetOrderQuery struct {
	Symbol            string
	OrderID           int64
	OrigClientOrderID string
}

// BinanceMyTradesQuery GET /api/v3/myTrades
type BinanceMyTradesQuery struct {
	Symbol    string
	OrderID   int64
	StartTime int64
	Limit     int
}

// BinanceOrder 下单/查单统一结构（字段随接口 subset）
type BinanceOrder struct {
	Symbol              string `json:"symbol"`
	OrderID             int64  `json:"orderId"`
	ClientOrderID       string `json:"clientOrderId"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status              string `json:"status"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	TimeInForce         string `json:"timeInForce"`
	Raw                 json.RawMessage
}

// BinanceTrade GET /api/v3/myTrades
type BinanceTrade struct {
	Symbol          string `json:"symbol"`
	ID              int64  `json:"id"`
	OrderID         int64  `json:"orderId"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	QuoteQty        string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
	Time            int64  `json:"time"`
}
