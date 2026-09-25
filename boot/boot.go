package boot

import (
	bot "github.com/qingpeng2016/ai-token-mall/application/bot"
	coreservice "github.com/qingpeng2016/ai-token-mall/application/core-service"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/http"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/http/alipay"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/mysql"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/redis"
	"github.com/qingpeng2016/ai-token-mall/interfaces/handler"
	"github.com/qingpeng2016/ai-token-mall/interfaces/rest"
	log2 "github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/notification"

	"go.uber.org/dig"
	"go.uber.org/zap/zapcore"
)

func init() {
	log2.NewLogger("ai-token-mall", "./log", zapcore.DebugLevel)
}

func BuildContainer() *dig.Container {
	c := dig.New()

	cfg := conf.NewCfg()
	_ = c.Provide(func() *conf.Config { return cfg })

	// HTTP
	_ = c.Provide(rest.NewRouter)
	_ = c.Provide(handler.NewUserHandler)
	_ = c.Provide(coreservice.NewUserService)
	_ = c.Provide(coreservice.NewAlipayService)
	_ = c.Provide(coreservice.NewBotScheduleConfigService)

	// Bot
	_ = c.Provide(bot.NewStatsJob)
	_ = c.Provide(bot.NewScheduler)
	_ = c.Provide(bot.NewEntry)

	// Infra
	_ = c.Provide(NewDBClient)
	_ = c.Provide(mysql.NewUserImpl)
	_ = c.Provide(mysql.NewStatsImpl)
	_ = c.Provide(mysql.NewBotScheduleConfigImpl)
	_ = c.Provide(redis.NewClient)
	_ = c.Provide(http.NewHTTPClient)
	_ = c.Provide(alipay.NewClient)

	_ = c.Provide(notification.NewNotificationManager)

	return c
}
