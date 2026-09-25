// Package binance 实现 Binance Spot REST；signed 请求 HMAC-SHA256(query)。
package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log2 "github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	repo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const defaultRecvWindow = 5000

type Binance struct {
	rc *resty.Client
}

func NewBinance() repo.BinanceRepo {
	rc := resty.New()
	rc.SetTimeout(conf.GetHTTPTimeout())
	return &Binance{rc: rc}
}

func baseURL(a *entity.BinanceAuth) string {
	if a != nil && a.BaseURL != "" {
		return strings.TrimRight(a.BaseURL, "/")
	}
	return conf.GetBinanceBaseURL()
}

func recvWindow() int64 {
	if w := conf.GetBinanceRecvWindow(); w > 0 {
		return w
	}
	return defaultRecvWindow
}

func maskKey(k string) string {
	if k == "" {
		return ""
	}
	if len(k) <= 8 {
		return "***"
	}
	return k[:4] + "..." + k[len(k)-4:]
}

func logReq(ctx context.Context, method, api, u string, fields ...zap.Field) {
	all := []zap.Field{
		zap.String("api", api),
		zap.String("http_method", method),
		zap.String("url", u),
	}
	all = append(all, fields...)
	log2.InfoZ(ctx, "binance api request", all...)
}

func logResp(ctx context.Context, api string, status int, body []byte) {
	log2.InfoZ(ctx, "binance api response",
		zap.String("api", api),
		zap.Int("http_status", status),
		zap.Int("body_bytes", len(body)),
		zap.String("response_body", string(body)),
	)
}

func checkAuth(a *entity.BinanceAuth) error {
	if a == nil || a.APIKey == "" || a.APISecret == "" {
		return fmt.Errorf("binance: APIKey and APISecret are required")
	}
	return nil
}

func sign(secret, query string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(query))
	return hex.EncodeToString(mac.Sum(nil))
}

func (b *Binance) Ping(ctx context.Context, auth *entity.BinanceAuth) error {
	u := baseURL(auth) + "/api/v3/ping"
	logReq(ctx, "GET", "Ping", u)
	resp, err := b.rc.R().Get(u)
	if err != nil {
		return err
	}
	logResp(ctx, "Ping", resp.StatusCode(), resp.Body())
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("binance ping: status %d", resp.StatusCode())
	}
	return nil
}

func (b *Binance) GetBookTicker(ctx context.Context, auth *entity.BinanceAuth, symbol string) (*entity.BinanceBookTicker, error) {
	if symbol == "" {
		return nil, fmt.Errorf("binance: symbol required")
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	u := baseURL(auth) + "/api/v3/ticker/bookTicker?" + params.Encode()
	logReq(ctx, "GET", "GetBookTicker", u, zap.String("symbol", symbol))
	resp, err := b.rc.R().Get(u)
	if err != nil {
		return nil, err
	}
	body := resp.Body()
	logResp(ctx, "GetBookTicker", resp.StatusCode(), body)
	if resp.StatusCode() != http.StatusOK {
		return nil, parseHTTPError(body, resp.StatusCode())
	}
	var t entity.BinanceBookTicker
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (b *Binance) CreateSpotOrder(ctx context.Context, auth entity.BinanceAuth, req entity.BinanceCreateOrderReq) (*entity.BinanceOrder, error) {
	if err := checkAuth(&auth); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("side", req.Side)
	params.Set("type", req.Type)
	if req.TimeInForce != "" {
		params.Set("timeInForce", req.TimeInForce)
	}
	params.Set("quantity", req.Quantity)
	params.Set("price", req.Price)
	if req.NewClientOrderID != "" {
		params.Set("newClientOrderId", req.NewClientOrderID)
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", strconv.FormatInt(recvWindow(), 10))
	query := params.Encode()
	params.Set("signature", sign(auth.APISecret, query))
	u := baseURL(&auth) + "/api/v3/order?" + params.Encode()
	logReq(ctx, "POST", "CreateSpotOrder", u,
		zap.String("api_key", maskKey(auth.APIKey)),
		zap.Any("order", req),
	)
	resp, err := b.rc.R().
		SetHeader("X-MBX-APIKEY", auth.APIKey).
		Post(u)
	if err != nil {
		return nil, err
	}
	body := resp.Body()
	logResp(ctx, "CreateSpotOrder", resp.StatusCode(), body)
	if resp.StatusCode() != http.StatusOK {
		return nil, parseHTTPError(body, resp.StatusCode())
	}
	return decodeOrder(body)
}

func (b *Binance) GetSpotOrder(ctx context.Context, auth entity.BinanceAuth, q entity.BinanceGetOrderQuery) (*entity.BinanceOrder, error) {
	if err := checkAuth(&auth); err != nil {
		return nil, err
	}
	if q.Symbol == "" {
		return nil, fmt.Errorf("binance: symbol required")
	}
	if q.OrderID == 0 && q.OrigClientOrderID == "" {
		return nil, fmt.Errorf("binance: orderId or origClientOrderId required")
	}
	params := url.Values{}
	params.Set("symbol", q.Symbol)
	if q.OrderID != 0 {
		params.Set("orderId", strconv.FormatInt(q.OrderID, 10))
	}
	if q.OrigClientOrderID != "" {
		params.Set("origClientOrderId", q.OrigClientOrderID)
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", strconv.FormatInt(recvWindow(), 10))
	query := params.Encode()
	params.Set("signature", sign(auth.APISecret, query))
	u := baseURL(&auth) + "/api/v3/order?" + params.Encode()
	logReq(ctx, "GET", "GetSpotOrder", u,
		zap.String("api_key", maskKey(auth.APIKey)),
		zap.Any("query", q),
	)
	resp, err := b.rc.R().
		SetHeader("X-MBX-APIKEY", auth.APIKey).
		Get(u)
	if err != nil {
		return nil, err
	}
	body := resp.Body()
	logResp(ctx, "GetSpotOrder", resp.StatusCode(), body)
	if resp.StatusCode() != http.StatusOK {
		return nil, parseHTTPError(body, resp.StatusCode())
	}
	return decodeOrder(body)
}

func (b *Binance) GetMyTrades(ctx context.Context, auth entity.BinanceAuth, q entity.BinanceMyTradesQuery) ([]entity.BinanceTrade, error) {
	if err := checkAuth(&auth); err != nil {
		return nil, err
	}
	if q.Symbol == "" {
		return nil, fmt.Errorf("binance: symbol required")
	}
	params := url.Values{}
	params.Set("symbol", q.Symbol)
	if q.OrderID != 0 {
		params.Set("orderId", strconv.FormatInt(q.OrderID, 10))
	}
	if q.StartTime > 0 {
		params.Set("startTime", strconv.FormatInt(q.StartTime, 10))
	}
	if q.Limit > 0 {
		params.Set("limit", strconv.Itoa(q.Limit))
	} else {
		params.Set("limit", "500")
	}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", strconv.FormatInt(recvWindow(), 10))
	query := params.Encode()
	params.Set("signature", sign(auth.APISecret, query))
	u := baseURL(&auth) + "/api/v3/myTrades?" + params.Encode()
	logReq(ctx, "GET", "GetMyTrades", u,
		zap.String("api_key", maskKey(auth.APIKey)),
		zap.Any("query", q),
	)
	resp, err := b.rc.R().
		SetHeader("X-MBX-APIKEY", auth.APIKey).
		Get(u)
	if err != nil {
		return nil, err
	}
	body := resp.Body()
	logResp(ctx, "GetMyTrades", resp.StatusCode(), body)
	if resp.StatusCode() != http.StatusOK {
		return nil, parseHTTPError(body, resp.StatusCode())
	}
	var trades []entity.BinanceTrade
	if err := json.Unmarshal(body, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}

func (b *Binance) GetDepth(ctx context.Context, auth *entity.BinanceAuth, symbol string, limit int) (*entity.BinanceDepth, error) {
	if symbol == "" {
		return nil, fmt.Errorf("binance: symbol required")
	}
	if limit <= 0 {
		limit = 100
	}
	u := fmt.Sprintf("%s/api/v3/depth?symbol=%s&limit=%d", baseURL(auth), symbol, limit)
	logReq(ctx, "GET", "GetDepth", u, zap.String("symbol", symbol), zap.Int("limit", limit))
	resp, err := b.rc.R().Get(u)
	if err != nil {
		return nil, err
	}
	body := resp.Body()
	logResp(ctx, "GetDepth", resp.StatusCode(), body)
	if resp.StatusCode() != http.StatusOK {
		return nil, parseHTTPError(body, resp.StatusCode())
	}
	var d entity.BinanceDepth
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (b *Binance) GetExchangeInfo(ctx context.Context, auth *entity.BinanceAuth, symbol string) (*entity.BinanceSymbolInfo, error) {
	if symbol == "" {
		return nil, fmt.Errorf("binance: symbol required")
	}
	u := fmt.Sprintf("%s/api/v3/exchangeInfo?symbol=%s", baseURL(auth), symbol)
	logReq(ctx, "GET", "GetExchangeInfo", u, zap.String("symbol", symbol))
	resp, err := b.rc.R().Get(u)
	if err != nil {
		return nil, err
	}
	body := resp.Body()
	logResp(ctx, "GetExchangeInfo", resp.StatusCode(), body)
	if resp.StatusCode() != http.StatusOK {
		return nil, parseHTTPError(body, resp.StatusCode())
	}
	var info entity.BinanceExchangeInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	for i := range info.Symbols {
		if info.Symbols[i].Symbol == symbol {
			return &info.Symbols[i], nil
		}
	}
	return nil, fmt.Errorf("binance: symbol %s not in exchangeInfo", symbol)
}

func (b *Binance) CancelSpotOrder(ctx context.Context, auth entity.BinanceAuth, symbol, origClientOrderID string) (*entity.BinanceOrder, error) {
	if err := checkAuth(&auth); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("origClientOrderId", origClientOrderID)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", strconv.FormatInt(recvWindow(), 10))
	query := params.Encode()
	params.Set("signature", sign(auth.APISecret, query))
	u := baseURL(&auth) + "/api/v3/order?" + params.Encode()
	logReq(ctx, "DELETE", "CancelSpotOrder", u, zap.String("origClientOrderId", origClientOrderID))
	resp, err := b.rc.R().SetHeader("X-MBX-APIKEY", auth.APIKey).Delete(u)
	if err != nil {
		return nil, err
	}
	body := resp.Body()
	logResp(ctx, "CancelSpotOrder", resp.StatusCode(), body)
	if resp.StatusCode() != http.StatusOK {
		return nil, parseHTTPError(body, resp.StatusCode())
	}
	return decodeOrder(body)
}

func decodeOrder(body []byte) (*entity.BinanceOrder, error) {
	var o entity.BinanceOrder
	if err := json.Unmarshal(body, &o); err != nil {
		return nil, err
	}
	o.Raw = append(json.RawMessage(nil), body...)
	return &o, nil
}

func parseHTTPError(body []byte, status int) error {
	var be entity.BinanceError
	if err := json.Unmarshal(body, &be); err == nil && be.Msg != "" {
		notification.SendErrorLog(context.Background(), "binance-api-error",
			zap.Int("http_status", status),
			zap.Int("code", be.Code),
			zap.String("msg", be.Msg))
		return &be
	}
	return fmt.Errorf("binance: http %d: %s", status, strings.TrimSpace(string(body)))
}
