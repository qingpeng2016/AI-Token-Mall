package handler

import (
	"net/http"

	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	ginresponse "github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/gin/response"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/errorx"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/rest/request"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TaskHandler struct {
	svc *coreservice.TaskService
}

func NewTaskHandler(svc *coreservice.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// CreateTask 触发强平任务
//
// @Summary      触发强平任务
// @Description  BeTrust 下发强平信号；liquidation_id 幂等。body 须含 timestamp、sign，见 docs/api.md 鉴权章节。
// @Tags         liquidation-task
// @Accept       json
// @Produce      json
// @Param        body      body    request.CreateTaskReq  true  "触发参数"
// @Success      200       {object}  ginresponse.ApiResp{data=response.CreateTaskResp}
// @Router       /liquidation/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req request.CreateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		notification.SendErrorLog(c, "task create bind failed", zap.Error(err))
		ginresponse.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	resp, status, err := h.svc.CreateTask(c, &req)
	if err != nil {
		notification.SendErrorLog(c, "task create failed", zap.Error(err))
		ginresponse.ResponseErr(c, errorx.ErrUnknown)
		return
	}
	c.JSON(status, ginresponse.ApiResp{Code: ginresponse.OkBizCode, Data: resp, Message: "success"})
}

// GetTask 查询强平任务进度与结算
//
// @Summary      查询强平任务
// @Tags         liquidation-task
// @Produce      json
// @Param        liquidation_id query   string  true  "强平任务 ID（参与 sign）"
// @Param        timestamp      query   int     true  "毫秒时间戳"
// @Param        sign           query   string  true  "HMAC 签名"
// @Success      200            {object}  ginresponse.ApiResp{data=response.GetTaskResp}
// @Router       /liquidation/tasks [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	liquidationID := c.Query("liquidation_id")
	if liquidationID == "" {
		ginresponse.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	resp, status, err := h.svc.GetTask(c, liquidationID)
	if err != nil {
		notification.SendErrorLog(c, "task get failed", zap.Error(err))
		ginresponse.ResponseErr(c, errorx.ErrUnknown)
		return
	}
	if status == http.StatusNotFound {
		ginresponse.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	c.JSON(status, ginresponse.ApiResp{Code: ginresponse.OkBizCode, Data: resp, Message: "success"})
}

// CancelTask 取消强平任务
//
// @Summary      取消强平任务
// @Description  停止后续卖出；已在途 IOC 由服务撤单
// @Tags         liquidation-task
// @Accept       json
// @Produce      json
// @Param        body  body  request.CancelTaskReq  true  "取消参数（含 liquidation_id、timestamp、sign）"
// @Success      200            {object}  ginresponse.ApiResp{data=response.CancelTaskResp}
// @Router       /liquidation/tasks/cancel [post]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	var req request.CancelTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		notification.SendErrorLog(c, "task cancel bind failed", zap.Error(err))
		ginresponse.ResponseErr(c, errorx.ErrParamsError)
		return
	}
	resp, status, err := h.svc.CancelTask(c, &req)
	if err != nil {
		notification.SendErrorLog(c, "task cancel failed", zap.Error(err))
		ginresponse.ResponseErr(c, errorx.ErrUnknown)
		return
	}
	c.JSON(status, ginresponse.ApiResp{Code: ginresponse.OkBizCode, Data: resp, Message: "success"})
}
