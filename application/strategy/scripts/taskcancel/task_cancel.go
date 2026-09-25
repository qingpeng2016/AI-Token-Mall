package taskcancel

import (
	"context"

	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	httpentity "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	httprepo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TaskCancel 处理 paused / cancelled 任务：查单对齐成交后撤销在途 IOC（不发新单）。
type TaskCancel struct {
	db          *gorm.DB
	taskRepo    repository.LiquidationTaskRepo
	orderRepo   repository.LiquidationOrderRepo
	binance     httprepo.BinanceRepo
	binanceSync common.BinanceSync
	platformSvc *coreservice.PlatformConfigService
}

// NewTaskCancel 注入依赖。
func NewTaskCancel(
	db *gorm.DB,
	taskRepo repository.LiquidationTaskRepo,
	orderRepo repository.LiquidationOrderRepo,
	binance httprepo.BinanceRepo,
	fillRepo repository.LiquidationTraderRepo,
	eventRepo repository.LiquidationTaskEventRepo,
	configRepo repository.LiquidationExecutionConfigRepo,
	platformSvc *coreservice.PlatformConfigService,
) *TaskCancel {
	return &TaskCancel{
		db: db, taskRepo: taskRepo, orderRepo: orderRepo, binance: binance,
		binanceSync: common.NewBinanceSync(db, binance, orderRepo, fillRepo, taskRepo, eventRepo, configRepo),
		platformSvc: platformSvc,
	}
}

// Run 扫描 paused/cancelled 任务，清理在途单。
func (j *TaskCancel) Run(ctx context.Context) {
	auth, err := j.platformSvc.BinanceSpotAuth(ctx)
	if err != nil {
		return
	}
	tasks, err := j.taskRepo.ListPausedOrCancelled(ctx)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-task_cancel-List", zap.Error(err))
		return
	}
	for _, task := range tasks {
		j.guardOne(ctx, auth, task)
	}
}

// guardOne 先 Sync 在途单成交，再统一 Cancel。
func (j *TaskCancel) guardOne(ctx context.Context, auth httpentity.BinanceAuth, task entity.LiquidationTask) {
	inFlight, err := j.orderRepo.HasInFlightByLiquidationID(ctx, task.LiquidationID)
	if err != nil {
		return
	}
	if !inFlight {
		common.MarkInFlightOrdersClearedIfDone(ctx, j.db, j.orderRepo, j.taskRepo, task.LiquidationID)
		return
	}
	orders, err := j.orderRepo.ListInFlightByLiquidationID(ctx, task.LiquidationID)
	if err != nil {
		return
	}
	taskCopy := task
	for i := range orders {
		row := &orders[i]
		if err = j.binanceSync.SyncOrder(ctx, auth, row, &taskCopy); err != nil {
			notification.SendErrorLog(ctx, "liquidation-task_cancel-Sync",
				zap.String("liquidation_id", task.LiquidationID),
				zap.String("client_order_id", row.ClientOrderID),
				zap.Error(err))
		}
	}
	common.CancelInFlightOrders(ctx, auth, j.db, j.binance, j.orderRepo, task.LiquidationID)
	for i := range orders {
		fresh, _ := j.orderRepo.FindByClientOrderID(ctx, orders[i].ClientOrderID)
		if fresh == nil {
			continue
		}
		if err = j.binanceSync.SyncOrder(ctx, auth, fresh, &taskCopy); err != nil {
			notification.SendErrorLog(ctx, "liquidation-task_cancel-SyncAfterCancel",
				zap.String("liquidation_id", task.LiquidationID),
				zap.String("client_order_id", fresh.ClientOrderID),
				zap.Error(err))
		}
	}
	common.MarkInFlightOrdersClearedIfDone(ctx, j.db, j.orderRepo, j.taskRepo, task.LiquidationID)
}
