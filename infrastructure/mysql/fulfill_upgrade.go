package mysql

import (
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

// fulfillUpgradeOrderPaid 升档订单支付成功：新订阅行 + 新 Key + 升档通知（履约与新购相同，通知模板不同）。
func fulfillUpgradeOrderPaid(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, payChannel string, now time.Time) error {
	sub, err := createSubscriptionForPurchase(tx, order, product, now)
	if err != nil {
		return err
	}
	apiKey, err := issueAPIKeyForSubscription(tx, order, sub, product, now)
	if err != nil {
		return err
	}
	if err := insertSubscriptionOrderNotification(tx, order, sub, apiKey, "subscription_upgraded", now); err != nil {
		return err
	}
	return insertOrderPayWalletFlow(tx, order, product, payChannel, now)
}
