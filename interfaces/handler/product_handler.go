package handler

import (
	coreservice "github.com/qingpeng2016/ai-token-mall/application/core-service"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productSvc *coreservice.ProductService
}

func NewProductHandler(productSvc *coreservice.ProductService) *ProductHandler {
	return &ProductHandler{productSvc: productSvc}
}

// List 在售商品列表（首页 catalog）
func (h *ProductHandler) List(c *gin.Context) {
	upstream := c.Query("sku_upstream_name")
	data, err := h.productSvc.List(c.Request.Context(), upstream)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// NavMenu 顶部套餐购买 mega 菜单 + 品牌下拉
func (h *ProductHandler) NavMenu(c *gin.Context) {
	data, err := h.productSvc.NavMenu(c.Request.Context())
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
