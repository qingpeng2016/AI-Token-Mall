package response

type CreateTaskResp struct {
	Result        string `json:"result"`         // accepted | rejected
	ResultMessage string `json:"result_message"` // 结果说明
}

type CancelTaskResp struct {
	CancelResult string `json:"cancel_result"` // accepted | already_finished | rejected
	Message      string `json:"message,omitempty"`
}

type GetTaskResp struct {
	LiquidationID          string     `json:"liquidation_id"`
	LoanOrderID            string     `json:"loan_order_id"`
	Status                 string     `json:"status"`
	Symbol                 string     `json:"symbol"`
	TotalQty               string     `json:"total_qty"`
	TotalAmount            string     `json:"total_amount"`
	TotalFee               string     `json:"total_fee"`
	TotalNetProceedsAmount string     `json:"total_net_proceeds_amount"`
	AvgPrice               string     `json:"avg_price"`
	RemainingDebtAmount    string     `json:"remaining_debt_amount"`
	RemainingCollateralQty string     `json:"remaining_collateral_qty"`
	CallbackStatus         *string    `json:"callback_status,omitempty"`
	Fills                  []FillItem `json:"fills"`
}

type FillItem struct {
	ExchangeTradeID string  `json:"exchange_trade_id"`
	BatchNo         string  `json:"batch_no"`
	Price           string  `json:"price"`
	Quantity        string  `json:"quantity"`
	Amount          string  `json:"amount"`
	Fee             string  `json:"fee"`
	TradeTimeMs     int64   `json:"trade_time_ms"`
	FeeAsset        *string `json:"fee_asset,omitempty"`
}
