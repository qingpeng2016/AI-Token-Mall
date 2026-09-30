package boot

import (
	bot "github.com/qingpeng2016/ai-token-mall/application/bot"
	subscriptionlifecycle "github.com/qingpeng2016/ai-token-mall/application/bot/scripts/subscription_lifecycle"
	alipaysvc "github.com/qingpeng2016/ai-token-mall/application/core-service/alipay"
	botscheduleconfig "github.com/qingpeng2016/ai-token-mall/application/core-service/bot_schedule_config"
	"github.com/qingpeng2016/ai-token-mall/application/core-service/enterprise"
	invoicesvc "github.com/qingpeng2016/ai-token-mall/application/core-service/invoice"
	notificationsvc "github.com/qingpeng2016/ai-token-mall/application/core-service/notification"
	"github.com/qingpeng2016/ai-token-mall/application/core-service/order"
	productsvc "github.com/qingpeng2016/ai-token-mall/application/core-service/product"
	"github.com/qingpeng2016/ai-token-mall/application/core-service/subscription"
	"github.com/qingpeng2016/ai-token-mall/application/core-service/tutorial"
	"github.com/qingpeng2016/ai-token-mall/application/core-service/user"
	log2 "github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/notification"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/http"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/http/invoicelookup"
	payinfra "github.com/qingpeng2016/ai-token-mall/infrastructure/http/alipay"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/mysql"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/redis"
	"github.com/qingpeng2016/ai-token-mall/interfaces/handler"
	"github.com/qingpeng2016/ai-token-mall/interfaces/rest"

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
	_ = c.Provide(handler.NewProductHandler)
	_ = c.Provide(handler.NewTutorialHandler)
	_ = c.Provide(handler.NewEnterpriseHandler)
	_ = c.Provide(handler.NewOrderHandler)
	_ = c.Provide(handler.NewSubscriptionHandler)
	_ = c.Provide(handler.NewInvoiceConfigHandler)
	_ = c.Provide(handler.NewUserNotificationHandler)
	_ = c.Provide(user.NewUserService)
	_ = c.Provide(notificationsvc.NewUserNotificationService)
	_ = c.Provide(invoicesvc.NewInvoiceConfigService)
	_ = c.Provide(invoicelookup.NewClient)
	_ = c.Provide(invoicesvc.NewEnterpriseLookup)
	_ = c.Provide(productsvc.NewProductService)
	_ = c.Provide(tutorial.NewTutorialService)
	_ = c.Provide(enterprise.NewEnterpriseService)
	_ = c.Provide(order.NewOrderFulfillService)
	_ = c.Provide(order.NewOrderService)
	_ = c.Provide(subscription.NewSubscriptionService)
	_ = c.Provide(alipaysvc.NewAlipayService)
	_ = c.Provide(botscheduleconfig.NewBotScheduleConfigService)

	// Bot
	_ = c.Provide(subscriptionlifecycle.NewSubscriptionLifecycleJob)
	_ = c.Provide(bot.NewScheduler)
	_ = c.Provide(bot.NewEntry)

	// Infra
	_ = c.Provide(NewDBClient)
	_ = c.Provide(mysql.NewUsersImpl)
	_ = c.Provide(mysql.NewProductsCategoryImpl)
	_ = c.Provide(mysql.NewProductsImpl)
	_ = c.Provide(mysql.NewTutorialCategoryImpl)
	_ = c.Provide(mysql.NewTutorialArticleImpl)
	_ = c.Provide(mysql.NewEnterpriseInquiryImpl)
	_ = c.Provide(mysql.NewEnterpriseProductsImpl)
	_ = c.Provide(mysql.NewUserOrdersImpl)
	_ = c.Provide(mysql.NewUserSubscriptionsImpl)
	_ = c.Provide(mysql.NewUserWalletFlowsImpl)
	_ = c.Provide(mysql.NewUserNotificationsImpl)
	_ = c.Provide(mysql.NewUserAPIKeysImpl)
	_ = c.Provide(mysql.NewPaymentCallbacksImpl)
	_ = c.Provide(mysql.NewUserInvoicesImpl)
	_ = c.Provide(mysql.NewUserInvoiceConfigImpl)
	_ = c.Provide(mysql.NewTransactorImpl)
	_ = c.Provide(mysql.NewBotScheduleConfigImpl)
	_ = c.Provide(redis.NewClient)
	_ = c.Provide(http.NewHTTPClient)
	_ = c.Provide(payinfra.NewClient)

	_ = c.Provide(notification.NewNotificationManager)

	return c
}
