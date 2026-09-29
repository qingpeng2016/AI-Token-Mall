package scripts

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/mysql"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	ModuleAITokenMall         = "ai_token_mall"
	TaskSubscriptionLifecycle = "subscription_lifecycle"
)

type SubscriptionLifecycleJob struct {
	db *gorm.DB
}

func NewSubscriptionLifecycleJob(db *gorm.DB) *SubscriptionLifecycleJob {
	return &SubscriptionLifecycleJob{db: db}
}

func (j *SubscriptionLifecycleJob) Run(ctx context.Context) {
	if err := mysql.RunSubscriptionLifecycle(ctx, j.db); err != nil {
		logger.ErrorZ(ctx, "subscription-lifecycle-run-failed", zap.Error(err))
		return
	}
	logger.InfoZ(ctx, "subscription-lifecycle-run-finished")
}
