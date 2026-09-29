package coreservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/constants"
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
	orderType := normalizeOrderType(req.OrderType)
	if orderType == "" {
		return nil, errorx.ErrParamsError
	}

	product, err := s.products.FindOnSaleByID(ctx, req.ProductID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if product == nil {
		return nil, errorx.ErrProductNotFound
	}

	userSubID, err := resolveOrderUserSubscriptionID(ctx, s.orders, userID, orderType, product.ID, req.UserSubscriptionID)
	if err != nil {
		return nil, err
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
		OrderType:          orderType,
		UserSubscriptionID: userSubID,
		Quantity:           qty,
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
		"order_type":            orderType,
		"user_subscription_id": userSubID,
		"quantity":              qty,
		"channel":            channel,
		"enterprise_invoice": req.EnterpriseInvoice,
	})
	order.PayChannel = channel
	order.OutTradeNo = outTradeNo
	order.RawRequestJSON = datatypes.JSON(rawReq)

	var createErr error
	switch orderType {
	case constants.OrderTypeRenewal:
		createErr = s.orders.CreateRenewalOrderReplacingPending(ctx, order)
	case constants.OrderTypeUpgrade:
		createErr = s.orders.CreateUpgradeOrderReplacingPending(ctx, order)
	case constants.OrderTypeQuotaAddon:
		createErr = s.orders.CreateQuotaAddonOrderReplacingPending(ctx, order)
	default:
		createErr = s.orders.CreateOrder(ctx, order)
	}
	if createErr != nil {
		return nil, errorx.ErrDbError
	}

	return createOrderRespFromEntity(order), nil
}

func createOrderRespFromEntity(order *entity.UserOrder) *response.CreateOrderResp {
	return &response.CreateOrderResp{
		OrderNo:            order.OrderNo,
		OutTradeNo:         order.OutTradeNo,
		OrderID:            order.ID,
		OrderType:          order.OrderType,
		UserSubscriptionID: order.UserSubscriptionID,
		Channel:            order.PayChannel,
		Status:             order.Status,
		TotalAmountCents:   order.TotalAmountCents,
		Currency:           order.Currency,
	}
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

	order, err := s.orders.FindOrderByOutTradeNo(ctx, req.OutTradeNo)
	if err != nil {
		return errorx.ErrDbError
	}
	if order == nil {
		return errorx.ErrPaymentNotFound
	}
	if order.PayChannel != channel {
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
		if strings.Contains(err.Error(), "renew: no active subscription") ||
			strings.Contains(err.Error(), "renew: subscription") && strings.Contains(err.Error(), "is not active") ||
			strings.Contains(err.Error(), "renew: subscription not found") ||
			strings.Contains(err.Error(), "upgrade: subscription not found") ||
			strings.Contains(err.Error(), "upgrade: missing user_subscription_id") ||
			strings.Contains(err.Error(), "quota_addon:") {
			return errorx.ErrRenewNoSubscription
		}
		if strings.Contains(err.Error(), "unknown order_type") {
			return errorx.ErrParamsError
		}
		return errorx.ErrDbError
	}
	return nil
}

func normalizeOrderType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", constants.OrderTypePurchase:
		return constants.OrderTypePurchase
	case constants.OrderTypeRenewal:
		return constants.OrderTypeRenewal
	case constants.OrderTypeUpgrade:
		return constants.OrderTypeUpgrade
	case constants.OrderTypeQuotaAddon:
		return constants.OrderTypeQuotaAddon
	default:
		return ""
	}
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
