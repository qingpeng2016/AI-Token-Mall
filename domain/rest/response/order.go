package response

type CreateOrderResp struct {
	OrderNo          string `json:"order_no"`
	OutTradeNo       string `json:"out_trade_no"`
	OrderID          uint   `json:"order_id"`
	OrderType          string `json:"order_type"`
	UserSubscriptionID uint   `json:"user_subscription_id"`
	Channel            string `json:"channel"`
	Status           string `json:"status"`
	TotalAmount Money `json:"total_amount"`
	Currency         string `json:"currency"`
}
