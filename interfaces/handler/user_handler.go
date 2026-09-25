package handler

import (
	coreservice "github.com/qingpeng2016/ai-token-mall/application/core-service"
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
		response.ResponseErr(c, err)
		return
	}
	data, err := h.userSvc.Register(c.Request.Context(), &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseErr(c, err)
		return
	}
	data, err := h.userSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.ResponseErr(c, err)
		return
	}
	response.ResponseSuccess(c, data)
}
