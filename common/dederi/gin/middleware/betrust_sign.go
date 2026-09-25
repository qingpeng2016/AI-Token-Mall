package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/betrustsign"
)

func abortSign(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    401,
		"message": message,
		"data":    nil,
	})
}

// BeTrustSignVerify 校验 BeTrust 请求体 sign + timestamp（路径参数可并入 extra）。
func BeTrustSignVerify(platform *coreservice.PlatformConfigService, pathParams ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret, err := platform.APISecret(c.Request.Context())
		if err != nil || secret == "" {
			abortSign(c, "platform config unavailable")
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			abortSign(c, "read body failed")
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var params map[string]string
		if len(bytes.TrimSpace(body)) > 0 {
			params, err = betrustsign.FlattenJSONObject(body)
			if err != nil {
				abortSign(c, "invalid json")
				return
			}
		} else {
			params = querySignParams(c)
		}
		for _, p := range pathParams {
			if v := c.Param(p); v != "" {
				params[p] = v
			}
		}

		if msg, ok := betrustsign.VerifyIncomingParams(secret, params, time.Now()); !ok {
			abortSign(c, msg)
			return
		}
		c.Next()
	}
}

func querySignParams(c *gin.Context) map[string]string {
	q := c.Request.URL.Query()
	out := make(map[string]string, len(q))
	for k, vals := range q {
		if len(vals) > 0 {
			out[k] = vals[0]
		}
	}
	return out
}
