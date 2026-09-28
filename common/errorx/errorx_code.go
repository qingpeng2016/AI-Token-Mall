package errorx

var (
	ErrParamsError   = NewRespErr(100100, "参数错误", "參數錯誤", "Params error.")
	ErrUnknown       = NewRespErr(100101, "未知错误，请联系客服", "未知錯誤，請聯繫客服", "Unknown error. Please contact customer.")
	ErrDbError       = NewRespErr(100109, "数据库错误", "資料庫錯誤", "Database error.")
	ErrUserExists    = NewRespErr(100201, "用户已存在", "用戶已存在", "User already exists.")
	ErrUserNotFound  = NewRespErr(100202, "用户不存在", "用戶不存在", "User not found.")
	ErrWrongPassword     = NewRespErr(100203, "密码错误", "密碼錯誤", "Wrong password.")
	ErrUserDisabled      = NewRespErr(100204, "账号已禁用", "帳號已禁用", "Account disabled.")
	ErrPasswordMismatch  = NewRespErr(100205, "两次密码不一致", "兩次密碼不一致", "Passwords do not match.")
	ErrInvalidCredential = NewRespErr(100206, "邮箱或手机号格式不正确", "郵箱或手機號格式不正確", "Invalid email or phone.")
	ErrProductNotFound   = NewRespErr(100301, "商品不存在或已下架", "商品不存在或已下架", "Product not found.")
	ErrOrderNotFound     = NewRespErr(100302, "订单不存在", "訂單不存在", "Order not found.")
	ErrPaymentNotFound   = NewRespErr(100303, "支付单不存在", "支付單不存在", "Payment not found.")
	ErrOrderNotPayable   = NewRespErr(100304, "订单状态不可支付", "訂單狀態不可支付", "Order is not payable.")
	ErrRenewNoSubscription = NewRespErr(100305, "没有可续费的套餐", "沒有可續費的套餐", "No active subscription to renew.")
)
