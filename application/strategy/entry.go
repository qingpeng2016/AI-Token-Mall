package strategy

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"

	"go.uber.org/zap"
)

// Entry 脚本侧入口：定时驱动 Scheduler。
type Entry struct {
	scheduler *Scheduler
	ns        *gocron.Scheduler
	startMu   sync.Mutex
}

func catch() {
	if e := recover(); e != nil {
		notification.SendErrorLog(context.Background(), "strategy panic", zap.Any("e", e), zap.Any("stack", debug.Stack()))
	}
}

func NewEntry(scheduler *Scheduler) *Entry {
	return &Entry{scheduler: scheduler}
}

func (e *Entry) Start() error {
	_ = notification.GetGlobalNotificationManager().SendInfoAlert(context.Background(), "策略脚本启动", []notification.FieldPair{
		{Key: "机器", Value: conf.Get_MACHINE_NAME()},
	})

	e.startMu.Lock()
	defer e.startMu.Unlock()
	if e.ns != nil {
		return nil
	}

	e.ns = gocron.NewScheduler(time.Local)
	_, err := e.ns.Every(10).Seconds().Name("Scheduler").WaitForSchedule().SingletonMode().Do(func() {
		defer catch()
		e.scheduler.Handle()
	})
	if err != nil {
		notification.SendErrorLog(context.Background(), "scheduler-job-create-failed", zap.Error(err))
		return err
	}
	e.ns.SetMaxConcurrentJobs(1, gocron.WaitMode)
	e.ns.StartAsync()

	logger.InfoZ(context.Background(), "strategy-started", zap.String("machine", conf.Get_MACHINE_NAME()))
	return nil
}

func (e *Entry) Stop() error {
	e.startMu.Lock()
	defer e.startMu.Unlock()
	if e.ns == nil {
		return nil
	}
	e.ns.Stop()
	e.ns = nil
	_ = e.scheduler.Stop()
	logger.InfoZ(context.Background(), "strategy-stopped")
	return nil
}
