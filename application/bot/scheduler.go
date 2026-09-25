package bot

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	coreservice "github.com/qingpeng2016/ai-token-mall/application/core-service"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/notification"
	entity2 "github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

type Scheduler struct {
	botScheduleConfigService *coreservice.BotScheduleConfigService
	statsJob                 *StatsJob
	ns                       *gocron.Scheduler
	jobMap                   map[string]*gocron.Job
	configMap                map[string]float64
	started                  bool
	mu                       sync.Mutex
}

func NewScheduler(
	botScheduleConfigService *coreservice.BotScheduleConfigService,
	statsJob *StatsJob,
) *Scheduler {
	return &Scheduler{
		botScheduleConfigService: botScheduleConfigService,
		statsJob:                 statsJob,
		ns:                       gocron.NewScheduler(time.Local),
		jobMap:                   make(map[string]*gocron.Job),
		configMap:                make(map[string]float64),
	}
}

func (s *Scheduler) Handle() {
	ctx := context.Background()
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			notification.SendErrorLog(ctx, "bot-scheduler-panic",
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	if !s.started {
		s.ns.StartAsync()
		s.started = true
	}

	where := map[string]interface{}{
		"module IN ?":              []string{ModuleAITokenMall},
		"is_enabled = ?":           1,
		"is_strategy_enabled = ?":  1,
	}
	_, configs, err := s.botScheduleConfigService.GetAllConfigs(ctx, where)
	if err != nil {
		notification.SendErrorLog(ctx, "bot-scheduler-load-config-failed", zap.Error(err))
		return
	}

	active := make(map[string]struct{})
	for _, cfg := range configs {
		key := jobKey(cfg)
		active[key] = struct{}{}
		if s.configMap[key] == cfg.IntervalSeconds && s.jobMap[key] != nil {
			continue
		}
		if old := s.jobMap[key]; old != nil {
			s.ns.RemoveByReference(old)
		}
		interval := cfg.IntervalSeconds
		if interval <= 0 {
			interval = 60
		}
		captured := cfg
		job, err := s.ns.Every(interval).Seconds().Name(key).WaitForSchedule().SingletonMode().Do(func() {
			s.runTask(context.Background(), captured)
		})
		if err != nil {
			logger.ErrorZ(ctx, "bot-scheduler-register-job-failed", zap.String("task", key), zap.Error(err))
			continue
		}
		s.jobMap[key] = job
		s.configMap[key] = interval
		logger.InfoZ(ctx, "bot-scheduler-job-registered", zap.String("task", key), zap.Float64("interval_sec", interval))
	}

	for key, job := range s.jobMap {
		if _, ok := active[key]; !ok {
			s.ns.RemoveByReference(job)
			delete(s.jobMap, key)
			delete(s.configMap, key)
		}
	}
}

func (s *Scheduler) runTask(ctx context.Context, cfg entity2.BotScheduleConfig) {
	switch cfg.TaskName {
	case TaskUserStats:
		s.statsJob.Run(ctx)
	default:
		logger.WarnZ(ctx, "bot-unknown-task", zap.String("module", cfg.Module), zap.String("task", cfg.TaskName))
	}
}

func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ns != nil {
		s.ns.Stop()
	}
	s.started = false
	s.jobMap = make(map[string]*gocron.Job)
	s.configMap = make(map[string]float64)
	return nil
}

func jobKey(cfg entity2.BotScheduleConfig) string {
	return cfg.Module + ":" + cfg.TaskName
}
