package mysql

import (
	"errors"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

// fulfillQuotaAddonOrderPaid 加购额度：active 订阅的 limit_tokens 增加对应商品的 limit_tokens。
func fulfillQuotaAddonOrderPaid(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, payChannel string, now time.Time) error {
	sub, err := loadSubscriptionForQuotaAddon(tx, order)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("quota_addon: subscription not found")
		}
		return err
	}
	if sub.Status != "active" {
		return fmt.Errorf("quota_addon: subscription %d is not active", sub.ID)
	}

	addon := product.LimitTokens
	if err := tx.Model(sub).Updates(map[string]interface{}{
		"limit_tokens": gorm.Expr("limit_tokens + ?", addon),
		"updated_at":   now,
	}).Error; err != nil {
		return err
	}
	if err := tx.First(sub, sub.ID).Error; err != nil {
		return err
	}
	if err := tx.Model(&entity.UserAPIKey{}).
		Where("user_subscription_id = ? AND status <> ?", sub.ID, "rotated").
		Updates(map[string]interface{}{
			"limit_tokens": gorm.Expr("limit_tokens + ?", addon),
			"updated_at":   now,
		}).Error; err != nil {
		return err
	}
	if err := insertSubscriptionOrderNotification(tx, order, sub, nil, "subscription_quota_added", now); err != nil {
		return err
	}

	return insertOrderPayWalletFlow(tx, order, product, payChannel, now)
}

func loadSubscriptionForQuotaAddon(tx *gorm.DB, order *entity.UserOrder) (*entity.UserSubscription, error) {
	if order.UserSubscriptionID == 0 {
		return nil, fmt.Errorf("quota_addon: missing user_subscription_id on order %s", order.OrderNo)
	}
	var sub entity.UserSubscription
	err := tx.Where("id = ? AND user_id = ?", order.UserSubscriptionID, order.UserID).First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}
