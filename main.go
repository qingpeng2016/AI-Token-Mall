// mojo-strategy-liquidation
//
// @title           mojo-strategy-liquidation API
// @version         1.0
// @description     质押清算策略服务（BeTrust 触发 / 查询 / 取消；鉴权见 docs/api.md）
// @BasePath        /
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gph-tech/fgmm-strategy-bitfinex/application/strategy"
	"github.com/gph-tech/fgmm-strategy-bitfinex/boot"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/constants"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/logger"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	_ "github.com/gph-tech/fgmm-strategy-bitfinex/docs"
	"github.com/gph-tech/fgmm-strategy-bitfinex/interfaces/rest"

	"go.uber.org/dig"
	"go.uber.org/zap"
)

func main() {
	flag.Parse()
	time.Local = constants.UtcLocation

	container := boot.BuildContainer()

	if conf.IsRunBot() {
		var function any = func(bot *strategy.Entry) error {
			go func() {
				if err := bot.Start(); err != nil {
					notification.SendErrorLog(context.Background(), "bot start failed", zap.Error(err))
				}
			}()
			return nil
		}
		if err := container.Invoke(function); err != nil {
			notification.SendErrorLog(context.Background(), "run bot failed", zap.String("error", err.Error()))
			panic(err)
		}
	} else {
		var function any = func(router *rest.Router) {
			router.Run(conf.GetServeConf())
		}
		if err := container.Invoke(function); err != nil {
			notification.SendErrorLog(context.Background(), "run http server failed", zap.String("error", err.Error()))
			panic(err)
		}
	}

	gracefulExit(container)
}

func gracefulExit(container *dig.Container) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logger.InfoZ(context.Background(), "graceful-shutdown-start")

	if conf.IsRunBot() {
		var function any = func(bot *strategy.Entry) error {
			return bot.Stop()
		}
		if err := container.Invoke(function); err != nil {
			logger.ErrorZ(context.Background(), "stop bot failed11", zap.String("error", err.Error()))
		}
	} else {
		var function any = func(router *rest.Router) {
			router.Close()
		}
		if err := container.Invoke(function); err != nil {
			notification.SendErrorLog(context.Background(), "stop http server failed", zap.String("error", err.Error()))
		}
	}

	logger.InfoZ(context.Background(), "graceful-shutdown-completed")
}
