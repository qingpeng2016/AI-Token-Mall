package request

type CreateOrderReq struct {
	ProductID         uint   `json:"product_id" binding:"required"`
	OrderType         string `json:"order_type" binding:"omitempty,oneof=purchase renewal"`
	Quantity          int    `json:"quantity" binding:"required,min=1,max=99"`
	Channel           string `json:"channel" binding:"required"`
	EnterpriseInvoice bool   `json:"enterprise_invoice"`
}

type PaymentNotifyReq struct {
	OutTradeNo   string `json:"out_trade_no" binding:"required"`
	TradeNo      string `json:"trade_no"`
	TradeStatus  string `json:"trade_status"`
}
