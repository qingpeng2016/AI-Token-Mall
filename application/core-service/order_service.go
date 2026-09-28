package coreservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"gorm.io/datatypes"
)

type OrderService struct {
	orders   repository.OrderRepo
	products repository.ProductRepo
}

func NewOrderService(orders repository.OrderRepo, products repository.ProductRepo) *OrderService {
	return &OrderService{orders: orders, products: products}
}

func (s *OrderService) CreateMockOrder(ctx context.Context, userID uint, req *request.CreateOrderReq) (*response.CreateOrderResp, error) {
	channel := normalizePayChannel(req.Channel)
	if channel == "" {
		return nil, errorx.ErrParamsError
	}

	product, err := s.products.FindOnSaleByID(ctx, req.ProductID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if product == nil {
		return nil, errorx.ErrProductNotFound
	}

	qty := req.Quantity
	subtotal := product.PriceCents * int64(qty)
	total := subtotal
	if req.EnterpriseInvoice {
		total += (subtotal * 6) / 100
	}

	now := time.Now()
	orderNo := fmt.Sprintf("AP%d%04d", now.Unix(), userID%10000)
	outTradeNo := fmt.Sprintf("PAY%d%d", now.UnixNano()/1e6, userID)

	order := &entity.UserOrder{
		OrderNo:           orderNo,
		UserID:            userID,
		ProductID:         product.ID,
		Quantity:          qty,
		UnitPriceCents:    product.PriceCents,
		Status:            "pending_payment",
		TotalAmountCents:  total,
		Currency:          product.Currency,
		EnterpriseInvoice: boolToInt(req.EnterpriseInvoice),
		ExpireAt:          now.Add(30 * time.Minute),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	rawReq, _ := json.Marshal(map[string]interface{}{
		"product_id":         product.ID,
		"quantity":           qty,
		"channel":            channel,
		"enterprise_invoice": req.EnterpriseInvoice,
	})
	payment := &entity.UserPayment{
		Channel:        channel,
		OutTradeNo:     outTradeNo,
		AmountCents:    total,
		Status:         "pending",
		RawRequestJSON: datatypes.JSON(rawReq),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.orders.CreateOrderWithPayment(ctx, order, payment); err != nil {
		return nil, errorx.ErrDbError
	}

	return &response.CreateOrderResp{
		OrderNo:          orderNo,
		OutTradeNo:       outTradeNo,
		OrderID:          order.ID,
		PaymentID:        payment.ID,
		Channel:          channel,
		Status:           order.Status,
		TotalAmountCents: total,
		Currency:         order.Currency,
	}, nil
}

func (s *OrderService) HandlePaymentNotify(ctx context.Context, channel string, req *request.PaymentNotifyReq) error {
	channel = normalizePayChannel(channel)
	if channel == "" || strings.TrimSpace(req.OutTradeNo) == "" {
		return errorx.ErrParamsError
	}
	status := strings.ToUpper(strings.TrimSpace(req.TradeStatus))
	if status != "" && status != "TRADE_SUCCESS" && status != "SUCCESS" {
		return errorx.ErrParamsError
	}

	payment, err := s.orders.FindPaymentByOutTradeNo(ctx, req.OutTradeNo)
	if err != nil {
		return errorx.ErrDbError
	}
	if payment == nil {
		return errorx.ErrPaymentNotFound
	}
	if payment.Channel != channel {
		return errorx.ErrParamsError
	}

	third := strings.TrimSpace(req.TradeNo)
	if third == "" {
		third = fmt.Sprintf("MOCK_%s_%d", channel, time.Now().UnixNano())
	}
	payload, _ := json.Marshal(req)
	idem := fmt.Sprintf("%s:%s", channel, req.OutTradeNo)

	in := repository.PaymentNotifyInput{
		Channel:        channel,
		OutTradeNo:     req.OutTradeNo,
		ThirdTradeNo:   third,
		IdempotencyKey: idem,
		PayloadJSON:    payload,
	}
	if err := s.orders.ApplyPaymentNotifySuccess(ctx, in); err != nil {
		if strings.Contains(err.Error(), "not payable") {
			return errorx.ErrOrderNotPayable
		}
		return errorx.ErrDbError
	}
	return nil
}

func normalizePayChannel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "alipay":
		return "alipay"
	case "wechat", "wxpay":
		return "wechat"
	case "paypal":
		return "paypal"
	default:
		return ""
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
