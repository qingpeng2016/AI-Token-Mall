package strategy

import (
	"context"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/callback"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/orderreconcile"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/taskcancel"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/taskalert"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/taskexec"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/tasktiercalc"
	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	entity2 "github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

const scheduleModule = "liquidation"

// Scheduler 按 bot_schedule_config 动态挂脚本任务
type Scheduler struct {
	botScheduleConfigService *coreservice.BotScheduleConfigService
	taskTierCalc             *tasktiercalc.TaskTierCalc
	taskAlert                *taskalert.TaskAlert
	taskExec                 *taskexec.TaskExec
	callback                 *callback.Callback
	orderReconcile           *orderreconcile.OrderReconcile
	taskCancel               *taskcancel.TaskCancel
	ns                       *gocron.Scheduler
	jobMap                   map[string]*gocron.Job
	configMap                map[string]float64
	started                  bool
	mu                       sync.Mutex
}

func NewScheduler(
	botScheduleConfigService *coreservice.BotScheduleConfigService,
	taskTierCalc *tasktiercalc.TaskTierCalc,
	taskAlert *taskalert.TaskAlert,
	taskExec *taskexec.TaskExec,
	callback *callback.Callback,
	orderReconcile *orderreconcile.OrderReconcile,
	taskCancel *taskcancel.TaskCancel,
) *Scheduler {
	return &Scheduler{
		botScheduleConfigService: botScheduleConfigService,
		taskTierCalc:             taskTierCalc,
		taskAlert:                taskAlert,
		taskExec:                 taskExec,
		callback:                 callback,
		orderReconcile:           orderReconcile,
		taskCancel:               taskCancel,
		ns:                       gocron.NewScheduler(time.Local),
		jobMap:                   make(map[string]*gocron.Job),
		configMap:                make(map[string]float64),
	}
}

func (s *Scheduler) Handle(configIDs ...uint) {
	ctx := context.Background()
	s.mu.Lock()
	machineName := conf.Get_MACHINE_NAME()
	logger.DebugZ(ctx, "scheduler-LockAcquired", zap.String("machineName", machineName))
	defer s.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			notification.SendErrorLog(ctx, "scheduler-Panic",
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	if !s.started {
		s.ns.StartAsync()
		s.started = true
		logger.InfoZ(ctx, "scheduler-Started")
	}

	where := map[string]interface{}{}
	hasConfigIDs := len(configIDs) > 0
	if hasConfigIDs {
		where["id IN ?"] = configIDs
	} else {
		where["module IN ?"] = []string{scheduleModule}
		where["is_enabled = ?"] = 1
		where["is_strategy_enabled = ?"] = 1
	}
	total, botConfigs, err := s.botScheduleConfigService.GetAllConfigs(ctx, where)
	if err != nil {
		notification.SendErrorLog(ctx, "scheduler-FailedToGetConfigs", zap.Error(err))
		return
	}
	configMap := make(map[string]entity2.BotScheduleConfig)
	for _, config := range botConfigs {
		configMap[config.Module+"-"+config.TaskName] = config
	}

	for jobName := range s.jobMap {
		if _, exists := configMap[jobName]; !exists {
			s.ns.RemoveByTags(jobName)
			delete(s.jobMap, jobName)
			delete(s.configMap, jobName)
			logger.InfoZ(ctx, "scheduler-JobRemoved", zap.String("jobName", jobName))
		}
	}
	if total == 0 || len(botConfigs) == 0 {
		logger.InfoZ(ctx, "scheduler-NoEnabledConfigs")
		return
	}

	for _, config := range botConfigs {
		handleFunc := s.getHandleFunc(config.Module, config.TaskName)
		if handleFunc == nil {
			notification.SendErrorLog(ctx, "scheduler-UnknownTask",
				zap.String("module", config.Module),
				zap.String("taskName", config.TaskName))
			continue
		}

		jobName := config.Module + "-" + config.TaskName
		module := config.Module
		taskName := config.TaskName
		needUpdate := false
		if _, exists := s.jobMap[jobName]; exists {
			if savedInterval, ok := s.configMap[jobName]; !ok || savedInterval != config.IntervalSeconds {
				needUpdate = true
			}
		}

		if !needUpdate && s.jobMap[jobName] != nil {
			continue
		}

		if needUpdate {
			s.ns.RemoveByTags(jobName)
			delete(s.jobMap, jobName)
			delete(s.configMap, jobName)
		}

		intervalDur := time.Duration(config.IntervalSeconds * float64(time.Second))
		if intervalDur < time.Millisecond {
			intervalDur = time.Millisecond
		}
		job, err := s.ns.Every(intervalDur).
			Tag(jobName).
			Name(jobName).
			WaitForSchedule().
			SingletonMode().
			Do(func() {
				defer catch()
				jobCtx := context.Background()
				if !hasConfigIDs {
					dbConfig, err := s.botScheduleConfigService.GetConfigByModuleAndTask(jobCtx, module, taskName)
					if err != nil {
						notification.SendErrorLog(jobCtx, "scheduler-FailedToGetConfig",
							zap.String("module", module),
							zap.String("taskName", taskName),
							zap.Error(err))
						return
					}
					if dbConfig == nil || dbConfig.IsEnabled != 1 {
						logger.InfoZ(jobCtx, "scheduler-TaskDisabled",
							zap.String("module", module),
							zap.String("taskName", taskName))
						return
					}
				}
				handleFunc()
			})
		if err != nil {
			notification.SendErrorLog(ctx, "scheduler-FailedToCreateJob",
				zap.String("module", config.Module),
				zap.String("taskName", config.TaskName),
				zap.Error(err))
			continue
		}

		s.jobMap[jobName] = job
		s.configMap[jobName] = config.IntervalSeconds
		if needUpdate {
			logger.InfoZ(ctx, "scheduler-JobUpdated",
				zap.String("jobName", jobName),
				zap.Float64("interval", config.IntervalSeconds))
		} else {
			logger.InfoZ(ctx, "scheduler-JobCreated",
				zap.String("jobName", jobName),
				zap.Float64("interval", config.IntervalSeconds))
		}
	}

	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
}

func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for jobName := range s.jobMap {
		_ = s.ns.RemoveByTags(jobName)
	}
	s.jobMap = make(map[string]*gocron.Job)
	s.configMap = make(map[string]float64)
	s.started = false
	s.ns.Stop()
	return nil
}

func (s *Scheduler) getHandleFunc(module, taskName string) func() {
	if module != scheduleModule {
		return nil
	}
	switch taskName {
	case "task_tier_calc":
		return func() { s.taskTierCalc.Run(context.Background()) }
	case "task_alert":
		return func() { s.taskAlert.Run(context.Background()) }
	case "task_exec":
		return func() { s.taskExec.Run(context.Background()) }
	case "result_callback":
		return func() { s.callback.Run(context.Background()) }
	case "order_updater":
		return func() { s.orderReconcile.Run(context.Background()) }
	case "task_cancel":
		return func() { s.taskCancel.Run(context.Background()) }
	default:
		return nil
	}
}
