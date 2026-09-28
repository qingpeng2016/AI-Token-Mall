package handler

import (
	"net/http"

	coreservice "github.com/qingpeng2016/ai-token-mall/application/core-service"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderSvc *coreservice.OrderService
}

func NewOrderHandler(orderSvc *coreservice.OrderService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc}
}

// CreateOrder 创建待支付订单（模拟收银台前置）
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.orderSvc.CreateMockOrder(c.Request.Context(), userID, &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// PaymentNotify 模拟微信/支付宝异步回调
func (h *OrderHandler) PaymentNotify(c *gin.Context) {
	channel := c.Param("channel")
	var req request.PaymentNotifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	if req.TradeStatus == "" {
		req.TradeStatus = "TRADE_SUCCESS"
	}
	if err := h.orderSvc.HandlePaymentNotify(c.Request.Context(), channel, &req); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, gin.H{"status": "success"})
}
