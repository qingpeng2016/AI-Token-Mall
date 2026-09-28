package mysql

import (
	"encoding/json"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// fulfillPurchaseOrderPaid 新购订单支付成功：新订阅行 + 新 Key + 开通通知。
func fulfillPurchaseOrderPaid(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, payChannel string, now time.Time) error {
	sub, err := createSubscriptionForPurchase(tx, order, product, now)
	if err != nil {
		return err
	}
	apiKey, err := issueAPIKeyForSubscription(tx, order, sub, product, now)
	if err != nil {
		return err
	}
	if err := insertSubscriptionOrderNotification(tx, order, sub, apiKey, "subscription_activated", now); err != nil {
		return err
	}
	return insertOrderPayWalletFlow(tx, order, product, payChannel, now)
}

func createSubscriptionForPurchase(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, now time.Time) (*entity.UserSubscription, error) {
	var oldSubs []entity.UserSubscription
	if err := tx.Where("user_id = ? AND product_id = ? AND status = ?", order.UserID, order.ProductID, "active").
		Find(&oldSubs).Error; err != nil {
		return nil, err
	}
	if len(oldSubs) > 0 {
		oldIDs := make([]uint, 0, len(oldSubs))
		for i := range oldSubs {
			oldIDs = append(oldIDs, oldSubs[i].ID)
		}
		if err := tx.Model(&entity.UserSubscription{}).Where("id IN ?", oldIDs).
			Updates(map[string]interface{}{"status": "expired", "updated_at": now}).Error; err != nil {
			return nil, err
		}
		if err := tx.Model(&entity.UserAPIKey{}).
			Where("user_subscription_id IN ? AND status = ?", oldIDs, "active").
			Updates(map[string]interface{}{
				"status":     "rotated",
				"rotated_at": now,
				"updated_at": now,
			}).Error; err != nil {
			return nil, err
		}
	}

	tokenGrant := product.LimitTokens * int64(order.Quantity)
	periodEnd := addBillingPeriod(now, product.BillingPeriod)
	ordersJSON, _ := json.Marshal([]uint{order.ID})

	row := entity.UserSubscription{
		UserID:               order.UserID,
		ProductID:            order.ProductID,
		Orders:               datatypes.JSON(ordersJSON),
		ProductsCategoryName: product.ProductsCategoryName,
		SKUProductName:       product.SKUProductName,
		BaseLimitTokens:      tokenGrant,
		LimitTokens:          tokenGrant,
		UsedTokens:           0,
		StartedAt:            now,
		ExpiresAt:            periodEnd,
		PeriodStart:          now,
		PeriodEnd:            periodEnd,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := tx.Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}
