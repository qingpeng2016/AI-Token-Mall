package boot

import (
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/callback"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/orderreconcile"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/taskcancel"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/taskalert"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/taskexec"
	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy/scripts/tasktiercalc"
	log2 "github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	"github.com/gph-tech/fgmm-strategy-bitfinex/infrastructure/http"
	"github.com/gph-tech/fgmm-strategy-bitfinex/infrastructure/http/binance"
	"github.com/gph-tech/fgmm-strategy-bitfinex/infrastructure/mysql"
	"github.com/gph-tech/fgmm-strategy-bitfinex/infrastructure/redis"
	"github.com/gph-tech/fgmm-strategy-bitfinex/interfaces/handler"
	"github.com/gph-tech/fgmm-strategy-bitfinex/interfaces/rest"

	"go.uber.org/dig"
	"go.uber.org/zap/zapcore"
)

func init() {
	log2.NewLogger("mojo", "./log", zapcore.DebugLevel)
}

func BuildContainer() *dig.Container {
	c := dig.New()

	cfg := conf.NewCfg()
	_ = c.Provide(func() *conf.Config {
		return cfg
	})

	// 接口侧
	_ = c.Provide(rest.NewRouter)
	_ = c.Provide(coreservice.NewPlatformConfigService)
	_ = c.Provide(coreservice.NewTaskService)
	_ = c.Provide(handler.NewTaskHandler)
	_ = c.Provide(handler.NewBeTrustContractHandler)

	// 脚本侧
	_ = c.Provide(tasktiercalc.NewTaskTierCalc)
	_ = c.Provide(taskalert.NewTaskAlert)
	_ = c.Provide(taskexec.NewTaskExec)
	_ = c.Provide(callback.NewCallback)
	_ = c.Provide(orderreconcile.NewOrderReconcile)
	_ = c.Provide(taskcancel.NewTaskCancel)
	_ = c.Provide(coreservice.NewBotScheduleConfigService)
	_ = c.Provide(strategy.NewScheduler)
	_ = c.Provide(strategy.NewEntry)

	// 基础设施（脚本侧读 bot_schedule_config 时才会连库）
	_ = c.Provide(NewDBClient)
	_ = c.Provide(mysql.NewBotScheduleConfigImpl)
	_ = c.Provide(mysql.NewLiquidationPlatformConfigImpl)
	_ = c.Provide(mysql.NewLiquidationInboundLogImpl)
	_ = c.Provide(mysql.NewLiquidationTaskImpl)
	_ = c.Provide(mysql.NewLiquidationCallbackLogImpl)
	_ = c.Provide(mysql.NewLiquidationOrderImpl)
	_ = c.Provide(mysql.NewLiquidationTraderImpl)
	_ = c.Provide(mysql.NewLiquidationExecutionConfigImpl)
	_ = c.Provide(mysql.NewLiquidationTaskEventImpl)
	_ = c.Provide(redis.NewClient)
	_ = c.Provide(http.NewHttpClient)
	_ = c.Provide(binance.NewBinance)

	_ = c.Provide(notification.NewNotificationManager)

	return c
}
