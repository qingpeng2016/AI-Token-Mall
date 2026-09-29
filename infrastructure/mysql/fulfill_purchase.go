package mysql

import (
	"encoding/json"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/apikey"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// fulfillPurchaseOrderPaid 新购订单支付成功：新增订阅行 + 新 Key（不改动历史订阅/Key）+ 开通通知。
func fulfillPurchaseOrderPaid(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, payChannel string, now time.Time) error {
	sub, err := createSubscriptionForPurchase(tx, order, product, now)
	if err != nil {
		return err
	}
	apiKey, err := createAPIKeyForNewSubscription(tx, order, sub, product, now)
	if err != nil {
		return err
	}
	if err := insertSubscriptionOrderNotification(tx, order, sub, apiKey, "subscription_activated", now); err != nil {
		return err
	}
	return insertOrderPayWalletFlow(tx, order, product, payChannel, now)
}

func createSubscriptionForPurchase(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, now time.Time) (*entity.UserSubscription, error) {
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

func createAPIKeyForNewSubscription(tx *gorm.DB, order *entity.UserOrder, sub *entity.UserSubscription, product *entity.Product, now time.Time) (*entity.UserAPIKey, error) {
	_, hash, err := apikey.Generate()
	if err != nil {
		return nil, err
	}
	row := entity.UserAPIKey{
		UserID:               order.UserID,
		UserSubscriptionID:   sub.ID,
		KeyHash:              hash,
		ProductsCategoryName: product.ProductsCategoryName,
		LimitTokens:          sub.LimitTokens,
		UsedTokens:           0,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := tx.Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}
