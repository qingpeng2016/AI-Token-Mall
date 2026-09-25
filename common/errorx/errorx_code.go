package errorx

var (
	ErrParamsError   = NewRespErr(100100, "参数错误", "參數錯誤", "Params error.")
	ErrUnknown       = NewRespErr(100101, "未知错误，请联系客服", "未知錯誤，請聯繫客服", "Unknown error. Please contact customer.")
	ErrDbError       = NewRespErr(100109, "数据库错误", "資料庫錯誤", "Database error.")
	ErrUserExists    = NewRespErr(100201, "用户已存在", "用戶已存在", "User already exists.")
	ErrUserNotFound  = NewRespErr(100202, "用户不存在", "用戶不存在", "User not found.")
	ErrWrongPassword = NewRespErr(100203, "密码错误", "密碼錯誤", "Wrong password.")
	ErrUserDisabled  = NewRespErr(100204, "账号已禁用", "帳號已禁用", "Account disabled.")
)
