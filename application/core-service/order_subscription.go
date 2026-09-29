package coreservice

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
)

func resolveOrderUserSubscriptionID(
	ctx context.Context,
	orders repository.OrderRepo,
	userID uint,
	orderType string,
	productID uint,
	reqSubID uint,
) (uint, error) {
	switch orderType {
	case constants.OrderTypePurchase:
		return 0, nil
	case constants.OrderTypeRenewal:
		sub, err := loadRenewalSubscription(ctx, orders, userID, productID, reqSubID)
		if err != nil {
			return 0, errorx.ErrDbError
		}
		if sub == nil {
			return 0, errorx.ErrRenewNoSubscription
		}
		return sub.ID, nil
	case constants.OrderTypeUpgrade:
		return resolveLinkedActiveSubscriptionID(ctx, orders, userID, reqSubID)
	case constants.OrderTypeQuotaAddon:
		return resolveLinkedActiveSubscriptionID(ctx, orders, userID, reqSubID)
	default:
		return 0, errorx.ErrParamsError
	}
}

// loadRenewalSubscription 续费前置：关联套餐必须为 active。
func loadRenewalSubscription(
	ctx context.Context,
	orders repository.OrderRepo,
	userID, productID, reqSubID uint,
) (*entity.UserSubscription, error) {
	if reqSubID > 0 {
		sub, err := orders.FindSubscriptionForUser(ctx, userID, reqSubID)
		if err != nil {
			return nil, err
		}
		if !subscriptionIsActive(sub) || sub.ProductID != productID {
			return nil, nil
		}
		return sub, nil
	}
	return orders.FindActiveSubscriptionByUserProduct(ctx, userID, productID)
}

func subscriptionIsActive(sub *entity.UserSubscription) bool {
	return sub != nil && sub.Status == "active"
}

// resolveLinkedActiveSubscriptionID 升档 / 加购额度：user_subscription_id 对应套餐必须为 active。
func resolveLinkedActiveSubscriptionID(ctx context.Context, orders repository.OrderRepo, userID, reqSubID uint) (uint, error) {
	if reqSubID == 0 {
		return 0, errorx.ErrParamsError
	}
	sub, err := orders.FindSubscriptionForUser(ctx, userID, reqSubID)
	if err != nil {
		return 0, errorx.ErrDbError
	}
	if !subscriptionIsActive(sub) {
		return 0, errorx.ErrRenewNoSubscription
	}
	return sub.ID, nil
}
