package handler

import (
	coreservice "github.com/qingpeng2016/ai-token-mall/application/core-service"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type EnterpriseHandler struct {
	enterpriseSvc *coreservice.EnterpriseService
}

func NewEnterpriseHandler(enterpriseSvc *coreservice.EnterpriseService) *EnterpriseHandler {
	return &EnterpriseHandler{enterpriseSvc: enterpriseSvc}
}

// SubmitInquiry 企业采购需求提交
func (h *EnterpriseHandler) SubmitInquiry(c *gin.Context) {
	var req request.SubmitEnterpriseInquiryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	id, err := h.enterpriseSvc.SubmitInquiry(c.Request.Context(), &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, gin.H{"id": id})
}

// ListProducts 企业方案列表
func (h *EnterpriseHandler) ListProducts(c *gin.Context) {
	data, err := h.enterpriseSvc.ListProducts(c.Request.Context())
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
