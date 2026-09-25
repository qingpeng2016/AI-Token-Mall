package common

import (
	"context"
	"strings"
	"time"

	httpentity "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	httprepo "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/repository"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"

	"gorm.io/gorm"
)

// CancelInFlightOrders 撤销本任务在 Binance 上仍标记为在途的子单（平台 cancel API / task_cancel 共用）。
func CancelInFlightOrders(ctx context.Context, auth httpentity.BinanceAuth, db *gorm.DB, binance httprepo.BinanceRepo, orderRepo repository.LiquidationOrderRepo, liquidationID string) {
	if binance == nil || orderRepo == nil || strings.TrimSpace(auth.APIKey) == "" {
		return
	}
	orders, err := orderRepo.ListInFlightByLiquidationID(ctx, liquidationID)
	if err != nil {
		return
	}
	now := time.Now()
	for _, o := range orders {
		// 撤单失败跳过，避免阻塞其余子单
		if _, err := binance.CancelSpotOrder(ctx, auth, o.Symbol, o.ClientOrderID); err != nil {
			continue
		}
		_ = orderRepo.Update(ctx, db, o.ID, map[string]interface{}{
			"status": entity.ExchangeOrderStatusCancelled, "finished_at": now,
		})
	}
}

// MarkInFlightOrdersClearedIfDone 无在途子单时置 in_flight_orders_cleared=1（task_cancel / 取消 API 共用）。
func MarkInFlightOrdersClearedIfDone(
	ctx context.Context,
	db *gorm.DB,
	orderRepo repository.LiquidationOrderRepo,
	taskRepo repository.LiquidationTaskRepo,
	liquidationID string,
) {
	if orderRepo == nil || taskRepo == nil {
		return
	}
	inFlight, err := orderRepo.HasInFlightByLiquidationID(ctx, liquidationID)
	if err != nil || inFlight {
		return
	}
	_ = taskRepo.UpdateByLiquidationID(ctx, db, liquidationID, map[string]interface{}{
		"in_flight_orders_cleared": entity.InFlightOrdersClearedYes,
	})
}
