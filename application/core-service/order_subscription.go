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
	case constants.OrderTypeUpgrade, constants.OrderTypeQuotaAddon:
		return resolveLinkedActiveSubscriptionID(ctx, orders, userID, reqSubID)
	default:
		return 0, errorx.ErrParamsError
	}
}

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
		if sub == nil || sub.Status != "active" || sub.ProductID != productID {
			return nil, nil
		}
		return sub, nil
	}
	return orders.FindActiveSubscriptionByUserProduct(ctx, userID, productID)
}

func resolveLinkedActiveSubscriptionID(ctx context.Context, orders repository.OrderRepo, userID, reqSubID uint) (uint, error) {
	if reqSubID == 0 {
		return 0, errorx.ErrParamsError
	}
	sub, err := orders.FindSubscriptionForUser(ctx, userID, reqSubID)
	if err != nil {
		return 0, errorx.ErrDbError
	}
	if sub == nil || sub.Status != "active" {
		return 0, errorx.ErrRenewNoSubscription
	}
	return sub.ID, nil
}
