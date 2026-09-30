package request

type RegisterUserReq struct {
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6"`
}

type LoginUserReq struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required"`
}

type ListWalletFlowsQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type ListInvoicesQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
