package request

type RegisterUserReq struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

type LoginUserReq struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required"`
}
