package taskalert

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/common"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"

	"go.uber.org/zap"
)

const alertCooldown = 5 * time.Minute

// TaskAlert 产品「内部告警」：同币对并发/名义、单任务 IOC 次数（只告警，不拦发单）。
type TaskAlert struct {
	taskRepo   repository.LiquidationTaskRepo
	orderRepo  repository.LiquidationOrderRepo
	configRepo repository.LiquidationExecutionConfigRepo
	dedupe     alertDedupe
}

// NewTaskAlert 注入依赖。
func NewTaskAlert(
	taskRepo repository.LiquidationTaskRepo,
	orderRepo repository.LiquidationOrderRepo,
	configRepo repository.LiquidationExecutionConfigRepo,
) *TaskAlert {
	return &TaskAlert{
		taskRepo:   taskRepo,
		orderRepo:  orderRepo,
		configRepo: configRepo,
		dedupe:     alertDedupe{m: make(map[string]time.Time)},
	}
}

// Run 扫描全部 liquidating 任务，按 symbol / 任务维度检查策略阈值。
func (j *TaskAlert) Run(ctx context.Context) {
	strategy, err := j.configRepo.GetStrategy(ctx)
	if err != nil || strategy == nil {
		return
	}

	tasks, err := j.taskRepo.ListRunnable(ctx, 0)
	if err != nil {
		notification.SendErrorLog(ctx, "liquidation-task_alert-ListRunnable", zap.Error(err))
		return
	}
	if len(tasks) == 0 {
		return
	}

	symbols := make(map[string]struct{})
	for _, t := range tasks {
		sym := common.ResolveSymbol(t)
		if sym != "" {
			symbols[sym] = struct{}{}
		}
	}

	for sym := range symbols {
		if strategy.MaxConcurrentTasksPerSymbol > 0 {
			n, err := j.taskRepo.CountLiquidatingBySymbol(ctx, sym)
			if err != nil {
				continue
			}
			if uint(n) > strategy.MaxConcurrentTasksPerSymbol {
				key := "concurrent_tasks:" + sym
				if j.dedupe.allow(key) {
					common.AlertLiquidation(ctx, "concurrent_tasks", "同币对在途强平过多", map[string]string{
						"symbol": sym,
						"count":  fmt.Sprintf("%d", n),
						"limit":  fmt.Sprintf("%d", strategy.MaxConcurrentTasksPerSymbol),
					})
				}
			}
		}
		if strategy.MaxNotionalUSDPerSymbol.IsPositive() {
			notional, err := common.SumLiquidatingNotionalBySymbol(ctx, j.taskRepo, j.orderRepo, sym)
			if err != nil {
				continue
			}
			if notional.GreaterThan(strategy.MaxNotionalUSDPerSymbol) {
				key := "notional_cap:" + sym
				if j.dedupe.allow(key) {
					common.AlertLiquidation(ctx, "notional_cap", "同币对在途强平名义金额超限", map[string]string{
						"symbol":   sym,
						"notional": notional.String(),
						"limit":    strategy.MaxNotionalUSDPerSymbol.String(),
					})
				}
			}
		}
	}

	if strategy.MaxOrdersPerTask > 0 {
		for _, t := range tasks {
			orderCount, err := j.orderRepo.CountByLiquidationID(ctx, t.LiquidationID)
			if err != nil {
				continue
			}
			if uint(orderCount) >= strategy.MaxOrdersPerTask {
				key := "max_orders:" + t.LiquidationID
				if j.dedupe.allow(key) {
					common.AlertLiquidation(ctx, "max_orders", "单任务 IOC 次数过多", map[string]string{
						"liquidation_id": t.LiquidationID,
						"order_count":    fmt.Sprintf("%d", orderCount),
						"limit":          fmt.Sprintf("%d", strategy.MaxOrdersPerTask),
					})
				}
			}
		}
	}
}

type alertDedupe struct {
	mu sync.Mutex
	m  map[string]time.Time
}

func (d *alertDedupe) allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	if t, ok := d.m[key]; ok && now.Sub(t) < alertCooldown {
		return false
	}
	d.m[key] = now
	return true
}
