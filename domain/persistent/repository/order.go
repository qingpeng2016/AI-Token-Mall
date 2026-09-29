package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type PaymentNotifyInput struct {
	Channel        string
	OutTradeNo     string
	ThirdTradeNo   string
	IdempotencyKey string
	PayloadJSON    []byte
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *entity.UserOrder) error
	// CreateRenewalOrderReplacingPending 取消同用户同订阅下 pending 的 renewal 单后新建（事务内）。
	CreateRenewalOrderReplacingPending(ctx context.Context, order *entity.UserOrder) error
	// CreateUpgradeOrderReplacingPending 取消同用户同订阅下 pending 的 upgrade 单后新建（事务内）。
	CreateUpgradeOrderReplacingPending(ctx context.Context, order *entity.UserOrder) error
	// CreateQuotaAddonOrderReplacingPending 取消同用户同订阅下 pending 的 quota_addon 单后新建（事务内）。
	CreateQuotaAddonOrderReplacingPending(ctx context.Context, order *entity.UserOrder) error
	FindOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.UserOrder, error)
	FindOrderByID(ctx context.Context, id uint) (*entity.UserOrder, error)
	FindActiveSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscription, error)
	FindLatestSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscription, error)
	FindSubscriptionForUser(ctx context.Context, userID, subscriptionID uint) (*entity.UserSubscription, error)
	ApplyPaymentNotifySuccess(ctx context.Context, in PaymentNotifyInput) error
}
