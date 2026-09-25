package betrustsign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MaxClockSkew = 10 * time.Second

// FlattenJSONObject 将 JSON 对象转为 sign 参与字段（sign 字段本身不参与）。
func FlattenJSONObject(body []byte) (map[string]string, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		s, err := rawMessageSignValue(v)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", k, err)
		}
		out[k] = s
	}
	return out, nil
}

func rawMessageSignValue(raw json.RawMessage) (string, error) {
	s := strings.TrimSpace(string(raw))
	if s == "null" {
		return "", nil
	}
	var v any
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return "", err
	}
	return canonicalValue(v), nil
}

func canonicalValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case float64:
		if x == float64(int64(x)) {
			return fmt.Sprintf("%d", int64(x))
		}
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", x), "0"), ".")
	case bool:
		if x {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		return fmt.Sprint(x)
	}
}

// BuildSignBase 按 key ASCII 升序拼接 key=value&...（不含 sign）。
func BuildSignBase(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	return b.String()
}

// ComputeSign hex(HMAC-SHA256(secret, base)) 小写。
func ComputeSign(secret, base string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(base))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySign(secret string, params map[string]string, providedSign string) bool {
	expect := ComputeSign(secret, BuildSignBase(params))
	return hmac.Equal([]byte(expect), []byte(providedSign))
}

func VerifyTimestamp(tsMillis int64, now time.Time) bool {
	if tsMillis <= 0 {
		return false
	}
	d := now.Sub(time.UnixMilli(tsMillis))
	if d < 0 {
		d = -d
	}
	return d <= MaxClockSkew
}

// VerifyIncomingParams 校验入站 timestamp + sign（params 为 body/query/path 合并后的扁平字段）。
func VerifyIncomingParams(secret string, params map[string]string, now time.Time) (message string, ok bool) {
	if secret == "" {
		return "app_secret not configured", false
	}
	sign := params["sign"]
	if sign == "" {
		return "missing sign", false
	}
	ts, err := strconv.ParseInt(params["timestamp"], 10, 64)
	if err != nil {
		return "invalid timestamp", false
	}
	if !VerifyTimestamp(ts, now) {
		return "request expired", false
	}
	if !VerifySign(secret, params, sign) {
		return "invalid sign", false
	}
	return "", true
}

// MarshalAndAttachSign 序列化 JSON 并按与验签相同的规则生成 sign（出站请求 BeTrust 等）。
func MarshalAndAttachSign(secret string, v any) ([]byte, error) {
	if secret == "" {
		return nil, fmt.Errorf("empty app_secret")
	}
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return AttachSign(secret, body)
}

// AttachSign 在 JSON 对象上写入 sign（先剔除旧 sign，再按文档重算）。
func AttachSign(secret string, body []byte) ([]byte, error) {
	params, err := FlattenJSONObject(body)
	if err != nil {
		return nil, err
	}
	delete(params, "sign")
	sig := ComputeSign(secret, BuildSignBase(params))

	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	m["sign"] = sig
	return json.Marshal(m)
}
