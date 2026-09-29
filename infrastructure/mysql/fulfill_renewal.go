package mysql

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// fulfillRenewalOrderPaid 续费：仅处理 active 订阅；不签发新 Key，只延长 expires_at；Key 恢复为 active。
func fulfillRenewalOrderPaid(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, payChannel string, now time.Time) error {
	sub, err := renewSubscriptionForOrder(tx, order, product, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("renew: subscription not found for order %s", order.OrderNo)
		}
		return err
	}
	if err := reactivateSubscriptionAPIKeys(tx, sub.ID, now); err != nil {
		return err
	}
	if err := insertSubscriptionOrderNotification(tx, order, sub, nil, "subscription_renewed", now); err != nil {
		return err
	}
	return insertOrderPayWalletFlow(tx, order, product, payChannel, now)
}

func renewSubscriptionForOrder(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, now time.Time) (*entity.UserSubscription, error) {
	sub, err := findSubscriptionForRenewal(tx, order)
	if err != nil {
		return nil, err
	}
	if sub.ProductID != order.ProductID {
		return nil, fmt.Errorf("renew: subscription %d product mismatch", sub.ID)
	}
	if sub.Status != "active" {
		return nil, fmt.Errorf("renew: subscription %d is not active", sub.ID)
	}

	var orderIDs []uint
	_ = json.Unmarshal(sub.Orders, &orderIDs)
	orderIDs = append(orderIDs, order.ID)
	ordersJSON, _ := json.Marshal(orderIDs)

	newExpires := addBillingPeriod(now, product.BillingPeriod)
	if sub.ExpiresAt.After(now) {
		newExpires = addBillingPeriod(sub.ExpiresAt, product.BillingPeriod)
	}

	updates := map[string]interface{}{
		"orders":     datatypes.JSON(ordersJSON),
		"expires_at": newExpires,
		"updated_at": now,
	}

	if err := tx.Model(sub).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := tx.First(sub, sub.ID).Error; err != nil {
		return nil, err
	}
	return sub, nil
}

func findSubscriptionForRenewal(tx *gorm.DB, order *entity.UserOrder) (*entity.UserSubscription, error) {
	var sub entity.UserSubscription
	var err error
	if order.UserSubscriptionID > 0 {
		err = tx.Where("id = ? AND user_id = ?", order.UserSubscriptionID, order.UserID).First(&sub).Error
	} else {
		err = tx.Where("user_id = ? AND product_id = ?", order.UserID, order.ProductID).
			Order("id DESC").First(&sub).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}
