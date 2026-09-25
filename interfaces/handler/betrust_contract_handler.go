package handler

import (
	"net/http"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/rest/request"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/rest/response"

	"github.com/gin-gonic/gin"
)

// BeTrustContractHandler 仅用于 Swagger 展示 BeTrust 需实现的回调契约（打桩，非生产逻辑）。
type BeTrustContractHandler struct{}

func NewBeTrustContractHandler() *BeTrustContractHandler {
	return &BeTrustContractHandler{}
}

// ReceiveLiquidationResult 强平结果回调（契约打桩）
//
// @Summary      接收强平结果回调（BeTrust 需实现）
// @Description  生产环境由本服务 callback 脚本 POST 到 BeTrust 配置的 callback_url；此路由仅在本服务 Swagger 中展示请求/响应格式，便于对方对接。
// @Tags         betrust-contract
// @Accept       json
// @Produce      json
// @Param        body  body  request.LiquidationTaskNotifyReq  true  "强平结果（与查询接口 data 同结构 + timestamp/sign）"
// @Success      200   {object}  response.BeTrustLiquidationCallbackResp
// @Router       /betrust/contract/liquidation/callback [post]
func (h *BeTrustContractHandler) ReceiveLiquidationResult(c *gin.Context) {
	var req request.LiquidationTaskNotifyReq
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, response.BeTrustLiquidationCallbackResp{
		Code:    200,
		Message: "stub: accepted",
	})
}
