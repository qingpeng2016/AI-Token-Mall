package request

type InviteRebatePageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type InvitePayoutConfigSaveReq struct {
	Channel string `json:"channel" binding:"required,oneof=alipay wechat"`
	QrURL   string `json:"qr_url" binding:"required,max=512"`
}
