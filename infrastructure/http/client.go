package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/trace"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	"github.com/go-resty/resty/v2"

	"go.uber.org/zap"
)

type Client struct {
	rc *resty.Client
}

func NewHttpClient() *Client {
	c := resty.New()
	c.SetTimeout(conf.GetHTTPTimeout())
	return &Client{rc: c}
}

func (c *Client) Get(ctx context.Context, api string, params map[string][]string, headers map[string]string) (*resty.Response, error) {
	reqParams := url.Values{}
	for k, v := range params {
		reqParams[k] = v
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers[trace.HeaderTraceID] = trace.GetTraceIdByCtx(ctx)

	// 构建查询字符串用于日志
	/*	queryString := reqParams.Encode()
		logger.InfoZ(ctx, "http get request",
			zap.String("api", api),
			zap.String("query", queryString))*/

	resp, err := c.rc.R().
		SetQueryParamsFromValues(reqParams).
		SetHeaders(headers).
		Get(api)

	if err != nil {
		notification.SendErrorLog(ctx, "http get error", zap.String("api", api), zap.String("error", err.Error()))
		return nil, err
	}

	// 记录响应状态码
	if resp.StatusCode() != 200 {
		// 如果响应体中有内容，打印出来
		if len(resp.Body()) > 0 {
			notification.SendErrorLog(ctx, "http get error with response body",
				zap.String("api", api),
				zap.Int("status_code", resp.StatusCode()),
				zap.String("response", string(resp.Body())))
		} else {
			notification.SendErrorLog(ctx, "http get error",
				zap.String("api", api),
				zap.Int("status_code", resp.StatusCode()))
		}
		return nil, fmt.Errorf("get %s error, status code: %d", api, resp.StatusCode())
	}

	return resp, err
}

func (c *Client) Post(ctx context.Context, api string, params map[string][]string, headers map[string]string, contentType string) (*resty.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[trace.HeaderTraceID] = trace.GetTraceIdByCtx(ctx)

	// 如果contentType是字符串且以{或[开头，尝试作为JSON字符串直接发送
	if contentType != "form" && contentType != "json" && (strings.HasPrefix(contentType, "{") || strings.HasPrefix(contentType, "[")) {
		headers["Content-Type"] = "application/json"

		// 记录请求
		logger.InfoZ(ctx, "http post direct json request",
			zap.String("api", api),
			zap.String("json", contentType))

		resp, err := c.rc.R().
			SetHeaders(headers).
			SetBody(contentType). // 直接使用字符串作为请求体
			Post(api)

		if err != nil {
			notification.SendErrorLog(ctx, "http post error", zap.String("api", api), zap.String("error", err.Error()))
			return nil, err
		}

		// 记录响应状态码
		if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
			// 如果响应体中有内容，打印出来
			if len(resp.Body()) > 0 {
				notification.SendErrorLog(ctx, "http post direct json error with response body",
					zap.String("api", api),
					zap.Int("status_code", resp.StatusCode()),
					zap.String("response", string(resp.Body())))
			} else {
				notification.SendErrorLog(ctx, "http post direct json error",
					zap.String("api", api),
					zap.Int("status_code", resp.StatusCode()))
			}
			return nil, fmt.Errorf("post %s error, status code: %d", api, resp.StatusCode())
		}

		return resp, err
	}

	if contentType == "form" {
		// 表单模式
		headers["Content-Type"] = "application/x-www-form-urlencoded"

		formData := url.Values{}
		for k, v := range params {
			for _, val := range v {
				formData.Add(k, val)
			}
		}

		// 构建表单数据字符串用于日志
		formString := formData.Encode()
		logger.InfoZ(ctx, "http post form request",
			zap.String("api", api),
			zap.String("form_data", formString))

		resp, err := c.rc.R().
			SetHeaders(headers).
			SetFormDataFromValues(formData).
			Post(api)

		if err != nil {
			notification.SendErrorLog(ctx, "http post error", zap.String("api", api), zap.String("error", err.Error()))
			return nil, err
		}

		// 记录响应状态码
		if resp.StatusCode() != 200 {
			// 如果响应体中有内容，打印出来
			if len(resp.Body()) > 0 {
				notification.SendErrorLog(ctx, "http post form error with response body",
					zap.String("api", api),
					zap.Int("status_code", resp.StatusCode()),
					zap.String("response", string(resp.Body())))
			} else {
				notification.SendErrorLog(ctx, "http post form error",
					zap.String("api", api),
					zap.Int("status_code", resp.StatusCode()))
			}
			return nil, fmt.Errorf("post %s error, status code: %d", api, resp.StatusCode())
		}

		return resp, err
	} else {
		// JSON 模式
		headers["Content-Type"] = "application/json"

		// 将 map[string][]string 转换为 map[string]interface{}
		jsonParams := make(map[string]interface{})
		for k, v := range params {
			if len(v) == 1 {
				// 特殊处理 wayParam，尝试解析为 JSON 对象
				if k == "wayParam" {
					var wayParamObj interface{}
					if err := json.Unmarshal([]byte(v[0]), &wayParamObj); err == nil {
						jsonParams[k] = wayParamObj
					} else {
						jsonParams[k] = v[0]
					}
				} else {
					jsonParams[k] = v[0]
				}
			} else {
				jsonParams[k] = v
			}
		}

		resp, err := c.rc.R().
			SetHeaders(headers).
			SetBody(jsonParams).
			Post(api)

		if err != nil {
			notification.SendErrorLog(ctx, "http post error", zap.String("api", api), zap.String("error", err.Error()))
			return nil, err
		}

		// 记录响应状态码
		if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
			// 如果响应体中有内容，打印出来
			if len(resp.Body()) > 0 {
				notification.SendErrorLog(ctx, "http post json error with response body",
					zap.String("api", api),
					zap.Int("status_code", resp.StatusCode()),
					zap.String("response", string(resp.Body())))
			} else {
				notification.SendErrorLog(ctx, "http post json error",
					zap.String("api", api),
					zap.Int("status_code", resp.StatusCode()))
			}
			return nil, fmt.Errorf("post %s error, status code: %d", api, resp.StatusCode())
		}

		return resp, err
	}
}

// Delete 发送 DELETE 请求，body 为 JSON 字符串
func (c *Client) Delete(ctx context.Context, api string, headers map[string]string, body string) (*resty.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[trace.HeaderTraceID] = trace.GetTraceIdByCtx(ctx)
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	resp, err := c.rc.R().
		SetHeaders(headers).
		SetBody(body).
		Delete(api)
	if err != nil {
		notification.SendErrorLog(ctx, "http delete error", zap.String("api", api), zap.String("error", err.Error()))
		return nil, err
	}
	if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
		if len(resp.Body()) > 0 {
			notification.SendErrorLog(ctx, "http delete error with response body",
				zap.String("api", api),
				zap.Int("status_code", resp.StatusCode()),
				zap.String("response", string(resp.Body())))
		} else {
			notification.SendErrorLog(ctx, "http delete error",
				zap.String("api", api),
				zap.Int("status_code", resp.StatusCode()))
		}
		return nil, fmt.Errorf("delete %s error, status code: %d", api, resp.StatusCode())
	}
	return resp, err
}
