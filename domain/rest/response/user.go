package response

type UserProfileResp struct {
	ID       uint   `json:"id"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Status   string `json:"status"`
}
