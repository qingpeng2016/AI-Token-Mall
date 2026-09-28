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
	FindOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.UserOrder, error)
	FindOrderByID(ctx context.Context, id uint) (*entity.UserOrder, error)
	FindActiveSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscription, error)
	FindSubscriptionForUser(ctx context.Context, userID, subscriptionID uint) (*entity.UserSubscription, error)
	ApplyPaymentNotifySuccess(ctx context.Context, in PaymentNotifyInput) error
}
