package handler

import (
	"net/http"

	inviteRebateSvc "github.com/qingpeng2016/ai-token-mall/application/core-service/invite_rebate"
	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type InviteRebateHandler struct {
	svc *inviteRebateSvc.Service
}

func NewInviteRebateHandler(svc *inviteRebateSvc.Service) *InviteRebateHandler {
	return &InviteRebateHandler{svc: svc}
}

func (h *InviteRebateHandler) Overview(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.GetOverview(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) ListMembers(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.ListMembers(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) ListCommissionRecords(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var q request.InviteRebatePageQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListCommissionRecords(c.Request.Context(), userID, &q)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) ListWithdrawals(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var q request.InviteRebatePageQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.ListWithdrawals(c.Request.Context(), userID, &q)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) GetPayoutConfig(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, err := h.svc.GetPayoutConfig(c.Request.Context(), userID)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

func (h *InviteRebateHandler) SavePayoutConfig(c *gin.Context) {
	userID, ok := ginMiddleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	var req request.InvitePayoutConfigSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	if err := h.svc.SavePayoutConfig(c.Request.Context(), userID, &req); err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, nil)
}
