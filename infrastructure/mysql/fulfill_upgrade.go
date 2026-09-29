package mysql

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/billing"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// fulfillUpgradeOrderPaid 升档：在原 user_subscription_id 上改 SKU/额度；active 不动周期与 expires；非 active 按商品重置周期并置 active。不签发新 Key。
func fulfillUpgradeOrderPaid(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, payChannel string, now time.Time) error {
	sub, wasActive, err := upgradeSubscriptionForOrder(tx, order, product, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("upgrade: subscription not found for order %s", order.OrderNo)
		}
		return err
	}
	if err := reactivateSubscriptionAPIKeys(tx, sub.ID, now); err != nil {
		return err
	}
	if err := syncAPIKeysAfterUpgrade(tx, sub.ID, product, sub.LimitTokens, wasActive, now); err != nil {
		return err
	}
	if err := insertSubscriptionOrderNotification(tx, order, sub, nil, "subscription_upgraded", now); err != nil {
		return err
	}
	return insertOrderPayWalletFlow(tx, order, product, payChannel, now)
}

func upgradeSubscriptionForOrder(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, now time.Time) (*entity.UserSubscription, bool, error) {
	if order.UserSubscriptionID == 0 {
		return nil, false, fmt.Errorf("upgrade: missing user_subscription_id on order %s", order.OrderNo)
	}

	var sub entity.UserSubscription
	if err := tx.Where("id = ? AND user_id = ?", order.UserSubscriptionID, order.UserID).First(&sub).Error; err != nil {
		return nil, false, err
	}

	tokenGrant := product.LimitTokens

	var orderIDs []uint
	_ = json.Unmarshal(sub.Orders, &orderIDs)
	orderIDs = append(orderIDs, order.ID)
	ordersJSON, _ := json.Marshal(orderIDs)

	wasActive := sub.Status == "active"

	updates := map[string]interface{}{
		"product_id":             product.ID,
		"orders":                 datatypes.JSON(ordersJSON),
		"products_category_name": product.ProductsCategoryName,
		"sku_product_name":       product.SKUProductName,
		"base_limit_tokens":      tokenGrant,
		"limit_tokens":           tokenGrant,
		"updated_at":             now,
	}

	if wasActive {
		// period_start / period_end / started_at / expires_at 保持不变
	} else {
		days := billing.PeriodDays(product.PeriodDays, product.BillingPeriod)
		qty := order.Quantity
		if qty < 1 {
			qty = 1
		}
		periodEnd := billing.AddPeriods(now, days, qty)
		updates["used_tokens"] = 0
		updates["started_at"] = now
		updates["expires_at"] = periodEnd
		updates["period_start"] = now
		updates["period_end"] = periodEnd
		updates["status"] = "active"
	}

	if err := tx.Model(&sub).Updates(updates).Error; err != nil {
		return nil, false, err
	}
	if err := tx.First(&sub, sub.ID).Error; err != nil {
		return nil, false, err
	}
	return &sub, wasActive, nil
}

func syncAPIKeysAfterUpgrade(tx *gorm.DB, subscriptionID uint, product *entity.Product, limitTokens int64, wasActiveBeforeUpgrade bool, now time.Time) error {
	keyUpdates := map[string]interface{}{
		"products_category_name": product.ProductsCategoryName,
		"limit_tokens":           limitTokens,
		"updated_at":             now,
	}
	if !wasActiveBeforeUpgrade {
		keyUpdates["used_tokens"] = 0
	}
	return tx.Model(&entity.UserAPIKey{}).
		Where("user_subscription_id = ? AND status <> ?", subscriptionID, "rotated").
		Updates(keyUpdates).Error
}
