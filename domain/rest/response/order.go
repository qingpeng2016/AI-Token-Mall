package response

type CreateOrderResp struct {
	OrderNo          string `json:"order_no"`
	OutTradeNo       string `json:"out_trade_no"`
	OrderID          uint   `json:"order_id"`
	PaymentID        uint   `json:"payment_id"`
	Channel          string `json:"channel"`
	Status           string `json:"status"`
	TotalAmountCents int64  `json:"total_amount_cents"`
	Currency         string `json:"currency"`
}
