package orderreconcile

import (
	"context"

	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	httprepo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const reconcileBatchSize = 50

// OrderReconcile 对 unknown / 超时 new 订单调用 Binance 查单并回写成交。
type OrderReconcile struct {
	db          *gorm.DB
	orderRepo   repository.LiquidationOrderRepo
	taskRepo    repository.LiquidationTaskRepo
	binance     httprepo.BinanceRepo
	binanceSync common.BinanceSync
	platformSvc *coreservice.PlatformConfigService
}

// NewOrderReconcile 注入依赖。
func NewOrderReconcile(
	db *gorm.DB,
	orderRepo repository.LiquidationOrderRepo,
	taskRepo repository.LiquidationTaskRepo,
	binance httprepo.BinanceRepo,
	fillRepo repository.LiquidationTraderRepo,
	eventRepo repository.LiquidationTaskEventRepo,
	configRepo repository.LiquidationExecutionConfigRepo,
	platformSvc *coreservice.PlatformConfigService,
) *OrderReconcile {
	return &OrderReconcile{
		db:          db,
		orderRepo:   orderRepo,
		taskRepo:    taskRepo,
		binance:     binance,
		binanceSync: common.NewBinanceSync(db, binance, orderRepo, fillRepo, taskRepo, eventRepo, configRepo),
		platformSvc: platformSvc,
	}
}

// Run 对 unknown / 超时 new 子单向 Binance 查单并回写。
func (j *OrderReconcile) Run(ctx context.Context) {
	auth, err := j.platformSvc.BinanceSpotAuth(ctx)
	if err != nil {
		return
	}

	orders, err := j.orderRepo.ListReconcileCandidates(ctx, reconcileBatchSize)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-order_reconcile-List", zap.Error(err))
		return
	}
	for i := range orders {
		row := &orders[i]
		task, _ := j.taskRepo.FindByLiquidationID(ctx, row.LiquidationID)
		if err = j.binanceSync.SyncOrder(ctx, auth, row, task); err != nil {
			notification.SendErrorLog(ctx, "liquidation-order_reconcile-Sync",
				zap.String("client_order_id", row.ClientOrderID),
				zap.Error(err))
			continue
		}
		if task != nil {
			taskAfter, _ := j.taskRepo.FindByLiquidationID(ctx, task.LiquidationID)
			if taskAfter != nil && taskAfter.Status == entity.LiquidationStatusLiquidating {
				_ = taskAfter
			}
		}
	}
}
