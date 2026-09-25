package callback

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/betrustsign"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/rest/request"
	httpclient "github.com/gph-tech/fgmm-strategy-bitfinex/infrastructure/http"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	callbackBatchSize   = 20
	callbackLarkBodyMax = 3500
)

// Callback 扫描终态且 callback 未成功的任务，POST BeTrust，写 liquidation_log_callback。
type Callback struct {
	db           *gorm.DB
	platformSvc  *coreservice.PlatformConfigService
	taskSvc      *coreservice.TaskService
	taskRepo     repository.LiquidationTaskRepo
	callbackRepo repository.LiquidationCallbackLogRepo
	http         *httpclient.Client
}

// NewCallback 注入依赖。
func NewCallback(
	db *gorm.DB,
	platformSvc *coreservice.PlatformConfigService,
	taskSvc *coreservice.TaskService,
	taskRepo repository.LiquidationTaskRepo,
	callbackRepo repository.LiquidationCallbackLogRepo,
	http *httpclient.Client,
) *Callback {
	return &Callback{
		db: db, platformSvc: platformSvc, taskSvc: taskSvc, taskRepo: taskRepo,
		callbackRepo: callbackRepo, http: http,
	}
}

// Run 批量拉取待回调任务并逐个 POST。
func (j *Callback) Run(ctx context.Context) {
	tasks, err := j.taskRepo.ListPendingCallback(ctx, callbackBatchSize)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-callback-ListPending", zap.Error(err))
		return
	}
	for _, task := range tasks {
		j.runOne(ctx, task)
	}
}

// shouldRetryNow 仅当已有成功回调日志时跳过；失败不更新任务态，调度周期内持续重试。
func (j *Callback) shouldRetryNow(ctx context.Context, liquidationID string) bool {
	latest, err := j.callbackRepo.LatestByLiquidationID(ctx, liquidationID)
	if err != nil || latest == nil {
		return true
	}
	return latest.Status != entity.CallbackLogStatusSuccess
}

// runOne 组装 payload、落 callback 日志、请求 BeTrust 并更新任务 callback 状态。
func (j *Callback) runOne(ctx context.Context, task entity.LiquidationTask) {
	callbackURL := strings.TrimSpace(task.CallbackURL)
	if callbackURL == "" {
		return
	}
	if !j.shouldRetryNow(ctx, task.LiquidationID) {
		return
	}

	now := time.Now()
	taskView, status, err := j.taskSvc.GetTask(ctx, task.LiquidationID)
	if err != nil || status != http.StatusOK || taskView == nil {
		notification.SendErrorLog(ctx, "liquidation-callback-BuildPayload",
			zap.String("liquidation_id", task.LiquidationID), zap.Error(err))
		return
	}
	payload := request.LiquidationTaskNotifyReq{
		Timestamp:   now.UnixMilli(),
		GetTaskResp: *taskView,
	}
	secret, err := j.platformSvc.APISecret(ctx)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-callback-Platform",
			zap.String("liquidation_id", task.LiquidationID), zap.Error(err))
		return
	}
	signedBody, err := betrustsign.MarshalAndAttachSign(secret, payload)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-callback-Sign",
			zap.String("liquidation_id", task.LiquidationID), zap.Error(err))
		return
	}

	logRow := &entity.LiquidationCallbackLog{
		LiquidationID: task.LiquidationID,
		CallbackURL:   callbackURL,
		Payload:       string(signedBody),
		Status:        entity.CallbackLogStatusPending,
	}
	if err := j.callbackRepo.Create(ctx, j.db, logRow); err != nil {
		notification.SendErrorLog(ctx, "liquidation-callback-LogCreate",
			zap.String("liquidation_id", task.LiquidationID),
			zap.Error(err))
		return
	}

	resp, postErr := j.http.Post(ctx, callbackURL, nil, nil, string(signedBody))
	now = time.Now()

	// 失败：只记 callback 日志，不更新 liquidation_task.callback_status，下一轮继续 POST
	if postErr != nil {
		httpStatus := 0
		respBody := postErr.Error()
		if resp != nil {
			httpStatus = resp.StatusCode()
			respBody = string(resp.Body())
		}
		_ = j.callbackRepo.Update(ctx, j.db, logRow.ID, map[string]interface{}{
			"status":        entity.CallbackLogStatusFailed,
			"http_status":   httpStatus,
			"response_body": respBody,
			"retry_count":   logRow.RetryCount + 1,
		})
		logger.ErrorZ(ctx, "liquidation-callback-PostFailed",
			zap.String("liquidation_id", task.LiquidationID),
			zap.Error(postErr))
		j.notifyCallbackToLark(ctx, task.LiquidationID, callbackURL, string(signedBody), httpStatus, respBody, false)
		return
	}

	httpStatus := resp.StatusCode()
	respBody := string(resp.Body())
	_ = j.callbackRepo.Update(ctx, j.db, logRow.ID, map[string]interface{}{
		"status":        entity.CallbackLogStatusSuccess,
		"http_status":   httpStatus,
		"response_body": respBody,
	})
	_ = j.taskRepo.UpdateByLiquidationID(ctx, j.db, task.LiquidationID, map[string]interface{}{
		"callback_status": entity.CallbackStatusSuccess,
		"callback_at":     now,
	})
	logger.InfoZ(ctx, "liquidation-callback-Success",
		zap.String("liquidation_id", task.LiquidationID),
		zap.Int("http_status", httpStatus))
	j.notifyCallbackToLark(ctx, task.LiquidationID, callbackURL, string(signedBody), httpStatus, respBody, true)
}

func (j *Callback) notifyCallbackToLark(ctx context.Context, liquidationID, callbackURL, requestBody string, httpStatus int, responseBody string, ok bool) {
	fields := []notification.FieldPair{
		{Key: "liquidation_id", Value: liquidationID},
		{Key: "callback_url", Value: callbackURL},
		{Key: "http_status", Value: strconv.Itoa(httpStatus)},
		{Key: "request_body", Value: truncateForLark(requestBody, callbackLarkBodyMax)},
		{Key: "response_body", Value: truncateForLark(responseBody, callbackLarkBodyMax)},
	}
	nm := notification.GetGlobalNotificationManager()
	if ok {
		_ = nm.SendInfoAlert(ctx, "清算终态回调", fields)
		return
	}
	_ = nm.SendErrorAlertWithOrder(ctx, "清算终态回调", fields)
}

func truncateForLark(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}
