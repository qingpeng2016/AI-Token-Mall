package response

// BeTrustLiquidationCallbackResp 建议 BeTrust 对回调的 HTTP 响应体（示例，具体以对方规范为准）。
type BeTrustLiquidationCallbackResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
