package request

import "github.com/gph-tech/fgmm-strategy-bitfinex/domain/rest/response"

// LiquidationTaskNotifyReq Mojo → BeTrust 终态通知：与 GET 查询接口的 data 同结构，并含 timestamp、sign。
type LiquidationTaskNotifyReq struct {
	Timestamp int64 `json:"timestamp"`
	Sign      string `json:"sign,omitempty"`
	response.GetTaskResp
}
