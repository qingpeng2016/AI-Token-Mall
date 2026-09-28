package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"

	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/notification"
	conf2 "github.com/qingpeng2016/ai-token-mall/conf"
	"github.com/qingpeng2016/ai-token-mall/interfaces/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	httpServer       *http.Server
	setting          *conf2.Config
	userHandler      *handler.UserHandler
	productHandler   *handler.ProductHandler
	tutorialHandler    *handler.TutorialHandler
	enterpriseHandler  *handler.EnterpriseHandler
}

func NewRouter(
	setting *conf2.Config,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	tutorialHandler *handler.TutorialHandler,
	enterpriseHandler *handler.EnterpriseHandler,
) *Router {
	return &Router{
		setting:           setting,
		userHandler:       userHandler,
		productHandler:    productHandler,
		tutorialHandler:   tutorialHandler,
		enterpriseHandler: enterpriseHandler,
	}
}

func (r *Router) setupRouters() *gin.Engine {
	engine := gin.Default()
	engine.Use(ginMiddleware.TraceRequestLog, ginMiddleware.CORSMiddleware)

	api := engine.Group("/api/v1")
	{
		api.POST("/users/register", r.userHandler.Register)
		api.POST("/users/login", r.userHandler.Login)
		api.POST("/users/logout", r.userHandler.Logout)
		api.GET("/products", r.productHandler.List)
		api.GET("/products/nav-menu", r.productHandler.NavMenu)
		api.GET("/products/slug/:slug", r.productHandler.DetailBySlug)
		api.GET("/tutorials", r.tutorialHandler.List)
		api.GET("/tutorials/articles/:slug", r.tutorialHandler.Detail)
		api.GET("/enterprise/products", r.enterpriseHandler.ListProducts)
		api.GET("/enterprise/products/:code", r.enterpriseHandler.GetProduct)
		api.POST("/enterprise/inquiries", r.enterpriseHandler.SubmitInquiry)
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
