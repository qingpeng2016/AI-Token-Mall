package bot

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/notification"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"go.uber.org/zap"
)

type Entry struct {
	scheduler *Scheduler
	ns        *gocron.Scheduler
	startMu   sync.Mutex
}

func NewEntry(scheduler *Scheduler) *Entry {
	return &Entry{scheduler: scheduler}
}

func (e *Entry) Start() error {
	_ = notification.GetGlobalNotificationManager().SendInfoAlert(context.Background(), "AI-Token-Mall bot 启动", []notification.FieldPair{
		{Key: "机器", Value: conf.Get_MACHINE_NAME()},
	})

	e.startMu.Lock()
	defer e.startMu.Unlock()
	if e.ns != nil {
		return nil
	}

	e.ns = gocron.NewScheduler(time.Local)
	_, err := e.ns.Every(10).Seconds().Name("BotScheduler").WaitForSchedule().SingletonMode().Do(func() {
		defer func() {
			if r := recover(); r != nil {
				notification.SendErrorLog(context.Background(), "bot-entry-panic", zap.Any("panic", r), zap.Any("stack", debug.Stack()))
			}
		}()
		e.scheduler.Handle()
	})
	if err != nil {
		return err
	}
	e.ns.SetMaxConcurrentJobs(1, gocron.WaitMode)
	e.ns.StartAsync()
	logger.InfoZ(context.Background(), "bot-started", zap.String("machine", conf.Get_MACHINE_NAME()))
	return nil
}

func (e *Entry) Stop() error {
	e.startMu.Lock()
	defer e.startMu.Unlock()
	if e.ns != nil {
		e.ns.Stop()
		e.ns = nil
	}
	_ = e.scheduler.Stop()
	logger.InfoZ(context.Background(), "bot-stopped")
	return nil
}
