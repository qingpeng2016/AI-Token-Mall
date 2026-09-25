package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"

	coreservice "github.com/gph-tech/fgmm-strategy-bitfinex/application/core-service"
	ginMiddleware "github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/gin/middleware"
	"github.com/gph-tech/fgmm-strategy-bitfinex/common/notification"
	conf2 "github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	"github.com/gph-tech/fgmm-strategy-bitfinex/interfaces/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Router struct {
	httpServer             *http.Server
	setting                *conf2.Config
	platformSvc            *coreservice.PlatformConfigService
	taskHandler            *handler.TaskHandler
	beTrustContractHandler *handler.BeTrustContractHandler
}

func NewRouter(
	setting *conf2.Config,
	platformSvc *coreservice.PlatformConfigService,
	taskHandler *handler.TaskHandler,
	beTrustContractHandler *handler.BeTrustContractHandler,
) *Router {
	return &Router{
		setting:                setting,
		platformSvc:            platformSvc,
		taskHandler:            taskHandler,
		beTrustContractHandler: beTrustContractHandler,
	}
}

func (r *Router) setupRouters() *gin.Engine {
	engine := gin.Default()
	engine.Use(ginMiddleware.TraceRequestLog, ginMiddleware.CORSMiddleware)
	engine.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))

	liqGroup := engine.Group("/liquidation")
	{
		liqGroup.POST("/tasks", ginMiddleware.BeTrustSignVerify(r.platformSvc), r.taskHandler.CreateTask)
		liqGroup.GET("/tasks", ginMiddleware.BeTrustSignVerify(r.platformSvc), r.taskHandler.GetTask)
		liqGroup.POST("/tasks/cancel", ginMiddleware.BeTrustSignVerify(r.platformSvc), r.taskHandler.CancelTask)
	}

	// 契约打桩：展示 BeTrust 需提供的强平结果回调（非生产入口）
	beTrustGroup := engine.Group("/betrust/contract")
	{
		beTrustGroup.POST("/liquidation/callback", r.beTrustContractHandler.ReceiveLiquidationResult)
	}

	return engine
}

func (r *Router) Run(setting *conf2.Server) {
	gin.SetMode(strings.ToLower(setting.RunMode))
	go func() {
		r.httpServer = &http.Server{Addr: setting.Port, Handler: r.setupRouters()}
		if err := r.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			notification.SendErrorLog(context.Background(), "run http server failed", zap.String("error", err.Error()))
			panic(err)
		}
	}()
}

func (r *Router) Close() {
	if r.httpServer == nil {
		return
	}
	if err := r.httpServer.Shutdown(context.Background()); err != nil {
		notification.SendErrorLog(context.Background(), "stop http server failed", zap.String("error", err.Error()))
	}
}
