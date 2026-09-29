package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qingpeng2016/ai-token-mall/common/auth"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
)

const ContextUserIDKey = "userID"

// RequireAuth 解析 Authorization: Bearer 或 Cookie atm_token
func RequireAuth(c *gin.Context) {
	token := ""
	if raw := strings.TrimSpace(c.GetHeader("Authorization")); raw != "" {
		token = strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
	}
	if token == "" {
		if cookie, err := c.Cookie(constants.AuthTokenCookie); err == nil {
			token = strings.TrimSpace(cookie)
		}
	}
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	uid, err := auth.ParseUserToken(token, auth.SessionTokenSecret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	c.Set(ContextUserIDKey, uid)
	c.Next()
}

func UserIDFromContext(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}
