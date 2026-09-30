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
	orderHandler          *handler.OrderHandler
	subscriptionHandler   *handler.SubscriptionHandler
	invoiceConfigHandler  *handler.InvoiceConfigHandler
}

func NewRouter(
	setting *conf2.Config,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	tutorialHandler *handler.TutorialHandler,
	enterpriseHandler *handler.EnterpriseHandler,
	orderHandler *handler.OrderHandler,
	subscriptionHandler *handler.SubscriptionHandler,
	invoiceConfigHandler *handler.InvoiceConfigHandler,
) *Router {
	return &Router{
		setting:              setting,
		userHandler:          userHandler,
		productHandler:       productHandler,
		tutorialHandler:      tutorialHandler,
		enterpriseHandler:    enterpriseHandler,
		orderHandler:         orderHandler,
		subscriptionHandler:  subscriptionHandler,
		invoiceConfigHandler: invoiceConfigHandler,
	}
}

func (r *Router) setupRouters() *gin.Engine {
	engine := gin.Default()
	engine.Use(ginMiddleware.TraceRequestLog, ginMiddleware.CORSMiddleware)

	v1 := engine.Group("/api/v1")
	{
		v1.POST("/users/register", r.userHandler.Register)
		v1.POST("/users/login", r.userHandler.Login)
		v1.POST("/users/logout", r.userHandler.Logout)
		v1.GET("/products", r.productHandler.List)
		v1.GET("/products/nav-menu", r.productHandler.NavMenu)
		v1.GET("/products/slug/:slug", r.productHandler.DetailBySlug)
		v1.GET("/tutorials", r.tutorialHandler.List)
		v1.GET("/tutorials/articles/:slug", r.tutorialHandler.Detail)
		v1.GET("/enterprise/products", r.enterpriseHandler.ListProducts)
		v1.GET("/enterprise/products/:code", r.enterpriseHandler.GetProduct)
		v1.POST("/enterprise/inquiries", r.enterpriseHandler.SubmitInquiry)
	}

	userAuth := engine.Group("/api/v1", ginMiddleware.RequireAuth)
	{
		userAuth.GET("/users/me", r.userHandler.Me)
		userAuth.GET("/users/wallet-flows", r.userHandler.ListWalletFlows)
		userAuth.GET("/users/invoices", r.userHandler.ListInvoices)
		userAuth.GET("/users/invoice-configs", r.invoiceConfigHandler.ListMine)
		userAuth.POST("/users/invoice-configs", r.invoiceConfigHandler.Create)
		userAuth.PUT("/users/invoice-configs/:id", r.invoiceConfigHandler.Update)
	}

	// 需登录；正式网关回调另开 /api/v1/payments/notify 且无鉴权
	authGroup := engine.Group("/api/v1/mock", ginMiddleware.RequireAuth)
	{
		authGroup.GET("/subscriptions", r.subscriptionHandler.ListMine)
		authGroup.GET("/orders", r.orderHandler.ListMine)
		authGroup.POST("/orders", r.orderHandler.CreateOrder)
		authGroup.POST("/orders/checkout", r.orderHandler.MockCheckout)
		authGroup.POST("/payments/notify/:channel", r.orderHandler.PaymentNotify)
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
