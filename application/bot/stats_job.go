package bot

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAITokenMall = "ai_token_mall"
	TaskUserStats     = "user_stats"
)

type StatsJob struct {
	stats repository.StatsRepo
}

func NewStatsJob(stats repository.StatsRepo) *StatsJob {
	return &StatsJob{stats: stats}
}

func (j *StatsJob) Run(ctx context.Context) {
	userCount, err := j.stats.CountUsers(ctx)
	if err != nil {
		logger.ErrorZ(ctx, "stats-count-users-failed", zap.Error(err))
		return
	}
	logCount, err := j.stats.CountAccessLogs(ctx)
	if err != nil {
		logger.ErrorZ(ctx, "stats-count-access-logs-failed", zap.Error(err))
		return
	}
	logger.InfoZ(ctx, "platform-stats",
		zap.Int64("user_count", userCount),
		zap.Int64("user_access_log_count", logCount),
	)
}
