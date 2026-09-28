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
	CreateOrderWithPayment(ctx context.Context, order *entity.UserOrder, payment *entity.UserPayment) error
	FindPaymentByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.UserPayment, error)
	FindOrderByID(ctx context.Context, id uint) (*entity.UserOrder, error)
	ApplyPaymentNotifySuccess(ctx context.Context, in PaymentNotifyInput) error
}
