package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type UserSubscriptionListRow struct {
	entity.UserSubscription
	ProductCardTitle string `gorm:"column:product_card_title"`
}

type SubscriptionRepo interface {
	ListByUserID(ctx context.Context, userID uint) ([]UserSubscriptionListRow, error)
}
