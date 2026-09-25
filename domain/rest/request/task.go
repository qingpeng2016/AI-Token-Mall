package request

type CreateTaskReq struct {
	Timestamp int64  `json:"timestamp" binding:"required" example:"1757923200123"` // 毫秒时间戳，参与 sign
	Sign      string `json:"sign" binding:"required" example:"a1b2..."`          // hex_lower(HMAC-SHA256(app_secret, 待签名字符串))
	LiquidationID    string `json:"liquidation_id" binding:"required"`
	LoanOrderID      string `json:"loan_order_id" binding:"required"`
	UID              string `json:"uid" binding:"required"`
	OriginalRatio    string `json:"original_ratio" binding:"required"`
	TriggerRatio     string `json:"trigger_ratio" binding:"required"`
	TriggerMarkPrice string `json:"trigger_mark_price" binding:"required"`
	TriggerCollateralAsset     string `json:"trigger_collateral_asset" binding:"required"`
	TriggerCollateralQty       string `json:"trigger_collateral_qty" binding:"required"`
	TriggerCollateralDueAmount string `json:"trigger_collateral_due_amount" binding:"required"`
	// Symbol Binance 现货交易对（执行所需，参与 sign）
	Symbol string `json:"symbol" binding:"required"`
}

type CancelTaskReq struct {
	LiquidationID string `json:"liquidation_id" binding:"required"`
	Timestamp     int64  `json:"timestamp" binding:"required" example:"1757923200123"`
	Sign          string `json:"sign" binding:"required" example:"a1b2..."`
	CancelReason  string `json:"cancel_reason" binding:"required"`
	RequestedAt   int64  `json:"requested_at" binding:"required"`
}
