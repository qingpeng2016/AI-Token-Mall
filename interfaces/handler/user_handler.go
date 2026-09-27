package handler

import (
	"net/http"

	coreservice "github.com/qingpeng2016/ai-token-mall/application/core-service"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/gin/response"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *coreservice.UserService
}

func NewUserHandler(userSvc *coreservice.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req request.RegisterUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.userSvc.Register(c.Request.Context(), &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		constants.AuthTokenCookie,
		data.Token,
		constants.AuthTokenMaxAge,
		"/",
		"",
		false,
		false,
	)
	response.ResponseSuccess(c, gin.H{"user": data.User})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseBindErr(c, err)
		return
	}
	data, err := h.userSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		constants.AuthTokenCookie,
		data.Token,
		constants.AuthTokenMaxAge,
		"/",
		"",
		false,
		false,
	)
	response.ResponseSuccess(c, gin.H{"user": data.User})
}

// Logout 退出登录：清空 atm_token Cookie
func (h *UserHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(constants.AuthTokenCookie, "", -1, "/", "", false, false)
	response.ResponseSuccess(c, nil)
}
