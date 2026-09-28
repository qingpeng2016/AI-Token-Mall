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

// fulfillRenewalOrderPaid 续费订单支付成功：更新原 active 订阅 + 轮换 Key + 续费通知。
func fulfillRenewalOrderPaid(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, payChannel string, now time.Time) error {
	sub, err := renewActiveSubscription(tx, order, product, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("renew: no active subscription for user %d product %d", order.UserID, order.ProductID)
		}
		return err
	}
	apiKey, err := issueAPIKeyForSubscription(tx, order, sub, product, now)
	if err != nil {
		return err
	}
	if err := insertSubscriptionOrderNotification(tx, order, sub, apiKey, "subscription_renewed", now); err != nil {
		return err
	}
	return insertOrderPayWalletFlow(tx, order, product, payChannel, now)
}

func renewActiveSubscription(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, now time.Time) (*entity.UserSubscription, error) {
	var sub entity.UserSubscription
	var err error
	if order.UserSubscriptionID > 0 {
		err = tx.Where("id = ? AND user_id = ? AND status = ?", order.UserSubscriptionID, order.UserID, "active").
			First(&sub).Error
	} else {
		err = tx.Where("user_id = ? AND product_id = ? AND status = ?", order.UserID, order.ProductID, "active").
			First(&sub).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	if sub.ProductID != order.ProductID {
		return nil, fmt.Errorf("renew: subscription %d product mismatch", sub.ID)
	}

	tokenGrant := product.LimitTokens * int64(order.Quantity)
	periodEnd := addBillingPeriod(now, product.BillingPeriod)

	var orderIDs []uint
	_ = json.Unmarshal(sub.Orders, &orderIDs)
	orderIDs = append(orderIDs, order.ID)
	ordersJSON, _ := json.Marshal(orderIDs)

	newExpires := periodEnd
	if sub.ExpiresAt.After(now) {
		newExpires = addBillingPeriod(sub.ExpiresAt, product.BillingPeriod)
	}

	if err := tx.Model(&sub).Updates(map[string]interface{}{
		"orders":                 datatypes.JSON(ordersJSON),
		"base_limit_tokens":      tokenGrant,
		"limit_tokens":           tokenGrant,
		"used_tokens":            0,
		"expires_at":             newExpires,
		"period_start":           now,
		"period_end":             addBillingPeriod(now, product.BillingPeriod),
		"products_category_name": product.ProductsCategoryName,
		"sku_product_name":       product.SKUProductName,
		"updated_at":             now,
	}).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}
