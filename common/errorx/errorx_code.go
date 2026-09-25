package errorx

var (
	ErrParamsError = NewRespErr(100100, "参数错误", "參數錯誤", "Params error.")
	ErrUnknown     = NewRespErr(100101, "未知错误，请联系客服", "未知錯誤，請聯繫客服", "Unknown error. Please contact customer.")
	ErrDbError     = NewRespErr(100109, "数据库错误", "資料庫錯誤", "Database error.")
)
