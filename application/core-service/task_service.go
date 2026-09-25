package coreservice

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	appcommon "github.com/gph-tech/fgmm-strategy-bitfinex/application/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	httprepo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/rest/request"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/rest/response"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	createTaskPath = "POST /liquidation/tasks"
	cancelTaskPath = "POST /liquidation/tasks/cancel"

	resultAccepted        = "accepted"         // 创建请求已受理
	resultRejected        = "rejected"         // 创建请求拒绝
	cancelAccepted        = "accepted"         // 取消已受理（异步撤单）
	cancelAlreadyFinished = "already_finished" // 任务已终态，无需取消
	cancelRejected        = "rejected"         // 取消拒绝
)

type TaskService struct {
	db          *gorm.DB
	platformSvc *PlatformConfigService
	inboundLog  repository.LiquidationInboundLogRepo
	taskRepo    repository.LiquidationTaskRepo
	orderRepo   repository.LiquidationOrderRepo
	fillRepo    repository.LiquidationTraderRepo
	eventRepo   repository.LiquidationTaskEventRepo
	binance     httprepo.BinanceRepo
}

func NewTaskService(
	db *gorm.DB,
	platformSvc *PlatformConfigService,
	inboundLog repository.LiquidationInboundLogRepo,
	taskRepo repository.LiquidationTaskRepo,
	orderRepo repository.LiquidationOrderRepo,
	fillRepo repository.LiquidationTraderRepo,
	eventRepo repository.LiquidationTaskEventRepo,
	binance httprepo.BinanceRepo,
) *TaskService {
	return &TaskService{
		db: db, platformSvc: platformSvc, inboundLog: inboundLog, taskRepo: taskRepo,
		orderRepo: orderRepo, fillRepo: fillRepo, eventRepo: eventRepo, binance: binance,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, req *request.CreateTaskReq) (*response.CreateTaskResp, int, error) {
	reqJSON, _ := json.Marshal(req)
	logID, err := s.appendInbound(ctx, entity.InboundAPIActionTrigger, createTaskPath, req.LiquidationID, req.LoanOrderID, req.Timestamp, string(reqJSON))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	resp := &response.CreateTaskResp{}
	status := http.StatusOK

	defer func() {
		s.finishInbound(ctx, logID, status, resp)
	}()

	existing, err := s.taskRepo.FindByLiquidationID(ctx, req.LiquidationID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if existing != nil {
		resp.Result = resultAccepted
		resp.ResultMessage = "强平任务已存在"
		return resp, status, nil
	}

	active, err := s.taskRepo.FindActiveByLoanOrderID(ctx, req.LoanOrderID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if active != nil && active.LiquidationID != req.LiquidationID {
		resp.Result = resultRejected
		resp.ResultMessage = "该借款订单已有进行中的强平任务"
		return resp, status, nil
	}

	originalRatio, err := decimal.NewFromString(req.OriginalRatio)
	if err != nil {
		resp.Result = resultRejected
		resp.ResultMessage = "original_ratio 无效"
		return resp, status, nil
	}
	triggerRatio, err := decimal.NewFromString(req.TriggerRatio)
	if err != nil {
		resp.Result = resultRejected
		resp.ResultMessage = "trigger_ratio 无效"
		return resp, status, nil
	}
	triggerMarkPrice, err := decimal.NewFromString(req.TriggerMarkPrice)
	if err != nil {
		resp.Result = resultRejected
		resp.ResultMessage = "trigger_mark_price 无效"
		return resp, status, nil
	}
	collateralQty, err := decimal.NewFromString(req.TriggerCollateralQty)
	if err != nil {
		resp.Result = resultRejected
		resp.ResultMessage = "trigger_collateral_qty 无效"
		return resp, status, nil
	}
	collateralDueAmount, err := decimal.NewFromString(req.TriggerCollateralDueAmount)
	if err != nil {
		resp.Result = resultRejected
		resp.ResultMessage = "trigger_collateral_due_amount 无效"
		return resp, status, nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	if symbol == "" {
		resp.Result = resultRejected
		resp.ResultMessage = "symbol 无效"
		return resp, status, nil
	}

	platform, err := s.platformSvc.GetBeTrust(ctx)
	if err != nil {
		resp.Result = resultRejected
		resp.ResultMessage = "平台配置不可用"
		return resp, status, nil
	}
	callbackURL := strings.TrimSpace(platform.CallbackURL)
	if callbackURL == "" {
		resp.Result = resultRejected
		resp.ResultMessage = "平台回调地址未配置"
		return resp, status, nil
	}

	now := time.Now()
	exchangeCode := entity.ExchangeCodeBinance
	feeRate := decimal.NewFromFloat(conf.GetBinanceFeeRate())
	tier := appcommon.LiquidationTierP2
	task := &entity.LiquidationTask{
		LiquidationID:              req.LiquidationID,
		LoanOrderID:                req.LoanOrderID,
		UID:                        req.UID,
		Status:                     entity.LiquidationStatusLiquidating,
		OriginalRatio:              originalRatio,
		TriggerRatio:               triggerRatio,
		TriggerMarkPrice:           triggerMarkPrice,
		TriggerCollateralAsset:     req.TriggerCollateralAsset,
		TriggerCollateralQty:       collateralQty,
		TriggerCollateralDueAmount: collateralDueAmount,
		Symbol:                     &symbol,
		CurrentTier:                &tier,
		FeeRate:                    feeRate,
		ExchangeCode:               &exchangeCode,
		CallbackURL:                callbackURL,
		StartedAt:                  &now,
	}

	if err := s.taskRepo.Create(ctx, s.db, task); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if s.eventRepo != nil {
		msg := "signal_received"
		_ = s.eventRepo.Create(ctx, s.db, &entity.LiquidationTaskEvent{
			LiquidationID: req.LiquidationID,
			EventType:     entity.TaskEventSignalReceived,
			ToStatus:      strPtr(entity.LiquidationStatusLiquidating),
			Message:       &msg,
		})
	}

	resp.Result = resultAccepted
	resp.ResultMessage = "强平任务已受理"
	return resp, status, nil
}

func (s *TaskService) CancelTask(ctx context.Context, req *request.CancelTaskReq) (*response.CancelTaskResp, int, error) {
	liquidationID := req.LiquidationID
	reqJSON, _ := json.Marshal(req)
	loanOrderID := ""
	if task, _ := s.taskRepo.FindByLiquidationID(ctx, liquidationID); task != nil {
		loanOrderID = task.LoanOrderID
	}

	logID, err := s.appendInbound(ctx, entity.InboundAPIActionCancel, cancelTaskPath, liquidationID, loanOrderID, req.RequestedAt, string(reqJSON))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	resp := &response.CancelTaskResp{}
	status := http.StatusOK

	defer func() {
		s.finishInbound(ctx, logID, status, resp)
	}()

	task, err := s.taskRepo.FindByLiquidationID(ctx, liquidationID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if task == nil {
		resp.CancelResult = cancelRejected
		resp.Message = "强平任务不存在"
		return resp, status, nil
	}

	switch task.Status {
	case entity.LiquidationStatusLiquidating, entity.LiquidationStatusPaused:
		// 执行中或已暂停均可取消
	default:
		resp.CancelResult = cancelAlreadyFinished
		resp.Message = "任务已终态，无法取消"
		return resp, status, nil
	}

	cancelAt := time.UnixMilli(req.RequestedAt)
	_ = s.cancelInFlightExchangeOrders(ctx, liquidationID)

	cancelResult := cancelAccepted
	endReason := entity.EndReasonCancelled
	updates := map[string]interface{}{
		"status":                   entity.LiquidationStatusCancelled,
		"end_reason":               endReason,
		"finished_at":              cancelAt,
		"cancel_reason":            req.CancelReason,
		"cancel_requested_at":      cancelAt,
		"cancel_result":            cancelResult,
		"in_flight_orders_cleared": entity.InFlightOrdersClearedNo,
	}
	if strings.TrimSpace(task.CallbackURL) != "" {
		updates["callback_status"] = entity.CallbackStatusPending
	}
	if err := s.taskRepo.UpdateCancel(ctx, s.db, liquidationID, updates); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	common.MarkInFlightOrdersClearedIfDone(ctx, s.db, s.orderRepo, s.taskRepo, liquidationID)

	resp.CancelResult = cancelAccepted
	resp.Message = "已受理，停止后续卖出"
	return resp, status, nil
}

func strPtr(s string) *string { return &s }

func (s *TaskService) GetTask(ctx context.Context, liquidationID string) (*response.GetTaskResp, int, error) {
	task, err := s.taskRepo.FindByLiquidationID(ctx, liquidationID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if task == nil {
		return nil, http.StatusNotFound, nil
	}
	symbol := ""
	if task.Symbol != nil {
		symbol = *task.Symbol
	}
	totals, err := common.TotalsFromTrades(ctx, s.fillRepo, liquidationID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	avgPrice := "0"
	if totals.Qty.IsPositive() {
		avgPrice = totals.Amount.Div(totals.Qty).String()
	}
	remainingDebtAmount := common.RemainingDebtAmount(*task, totals)
	remainingCollateralQty := common.RemainingCollateralQty(*task, totals)
	resp := &response.GetTaskResp{
		LiquidationID:          task.LiquidationID,
		LoanOrderID:            task.LoanOrderID,
		Status:                 task.Status,
		Symbol:                 symbol,
		TotalQty:               totals.Qty.String(),
		TotalAmount:            totals.Amount.String(),
		TotalFee:               totals.Fee.String(),
		TotalNetProceedsAmount: totals.Net.String(),
		AvgPrice:               avgPrice,
		RemainingDebtAmount:    remainingDebtAmount.String(),
		RemainingCollateralQty: remainingCollateralQty.String(),
		CallbackStatus:         task.CallbackStatus,
		Fills:                  []response.FillItem{},
	}
	if s.fillRepo != nil {
		fills, _ := s.fillRepo.ListByLiquidationID(ctx, liquidationID)
		for _, f := range fills {
			resp.Fills = append(resp.Fills, response.FillItem{
				ExchangeTradeID: f.ExchangeTradeID,
				BatchNo:         f.BatchNo,
				Price:           f.Price.String(),
				Quantity:        f.Quantity.String(),
				Amount:          f.Amount.String(),
				Fee:             f.Fee.String(),
				TradeTimeMs:     f.TradeTime.UnixMilli(),
				FeeAsset:        f.FeeAsset,
			})
		}
	}
	return resp, http.StatusOK, nil
}

func (s *TaskService) cancelInFlightExchangeOrders(ctx context.Context, liquidationID string) error {
	auth, err := s.platformSvc.BinanceSpotAuth(ctx)
	if err != nil {
		return err
	}
	common.CancelInFlightOrders(ctx, auth, s.db, s.binance, s.orderRepo, liquidationID)
	return nil
}

func (s *TaskService) appendInbound(ctx context.Context, action, apiPath, liquidationID, loanOrderID string, eventTimeMs int64, requestBody string) (uint, error) {
	row := &entity.LiquidationInboundLog{
		LiquidationID:  liquidationID,
		LoanOrderID:    loanOrderID,
		APIPath:        apiPath,
		APIAction:      action,
		EventTimeMs:    eventTimeMs,
		IdempotencyKey: liquidationID,
		RequestBody:    requestBody,
	}
	if err := s.inboundLog.Create(ctx, s.db, row); err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (s *TaskService) finishInbound(ctx context.Context, logID uint, httpStatus int, resp any) {
	body, _ := json.Marshal(resp)
	_ = s.inboundLog.UpdateResponse(ctx, s.db, logID, httpStatus, string(body))
}
