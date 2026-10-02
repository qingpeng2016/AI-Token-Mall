package apikey

import "strings"

// PrefixFromPlaintext 入库展示用前缀（ap- + 前 6 位随机段）。
func PrefixFromPlaintext(plaintext string) string {
	plaintext = strings.TrimSpace(plaintext)
	if len(plaintext) <= 10 {
		return plaintext
	}
	return plaintext[:10]
}

// Masked 脱敏展示。
func Masked(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "ap-••••••••••••"
	}
	if len(prefix) >= 10 {
		return prefix[:7] + "••••••" + prefix[len(prefix)-2:]
	}
	return prefix + "••••••"
}
