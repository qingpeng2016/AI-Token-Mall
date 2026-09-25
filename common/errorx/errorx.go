package errorx

import "strings"

const (
	LocaleHans = "zh-Hans"
	LocaleHant = "zh-Hant"
	LocaleEn   = "en"
)

// RespErr 自定义响应错误（简中 / 繁中 / 英文）
type RespErr struct {
	Code    int64       `json:"code"`               // 错误码
	Msg     string      `json:"msg"`                // 简体中文
	HantMsg string      `json:"hant_msg,omitempty"` // 繁体中文
	EnMsg   string      `json:"en_msg"`             // 英文
	Data    interface{} `json:"data,omitempty"`
}

func (e *RespErr) Error() string {
	return e.Msg
}

// NewRespErr 创建三语业务错误；hant 为空时回退为 zh。
func NewRespErr(code int64, zh, hant, en string) *RespErr {
	if hant == "" {
		hant = zh
	}
	return &RespErr{
		Code:    code,
		Msg:     zh,
		HantMsg: hant,
		EnMsg:   en,
	}
}

// NormalizeLocale 规范为 zh-Hans / zh-Hant / en。
func NormalizeLocale(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return LocaleHans
	}
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "zh-hant"), strings.HasPrefix(lower, "zh-tw"), strings.HasPrefix(lower, "zh-hk"):
		return LocaleHant
	case strings.HasPrefix(lower, "en"):
		return LocaleEn
	case strings.HasPrefix(lower, "zh"):
		return LocaleHans
	default:
		return LocaleHans
	}
}

// LocalizedMessage 按界面语言返回提示文案。
func (e *RespErr) LocalizedMessage(locale string) string {
	switch NormalizeLocale(locale) {
	case LocaleHant:
		if e.HantMsg != "" {
			return e.HantMsg
		}
		return e.Msg
	case LocaleEn:
		if e.EnMsg != "" {
			return e.EnMsg
		}
		return e.Msg
	default:
		return e.Msg
	}
}
