package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qingpeng2016/ai-token-mall/common/auth"
)

const ContextUserIDKey = "userID"

// RequireAuth 解析 Authorization: Bearer 登录 token
func RequireAuth(c *gin.Context) {
	raw := strings.TrimSpace(c.GetHeader("Authorization"))
	if raw == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	token := strings.TrimPrefix(raw, "Bearer ")
	token = strings.TrimSpace(token)
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
