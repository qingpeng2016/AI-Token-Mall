package mysql

import (
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// WalletBalanceSum Bot 脚本：用户钱包余额（流水 SUM）。
func WalletBalanceSum(tx *gorm.DB, userID uint) (decimal.Decimal, error) {
	return userWalletBalanceSum(tx, userID)
}

// BotRenewSubscriptionForOrder Bot 自动续费：延长订阅 expires_at。
func BotRenewSubscriptionForOrder(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, now time.Time) (*entity.UserSubscription, error) {
	return renewSubscriptionForOrder(tx, order, product, now)
}

// BotReactivateSubscriptionAPIKeys Bot 续费后恢复 Key 为 active。
func BotReactivateSubscriptionAPIKeys(tx *gorm.DB, subscriptionID uint, now time.Time) error {
	return reactivateSubscriptionAPIKeys(tx, subscriptionID, now)
}

// BotInsertSubscriptionOrderNotification Bot 脚本写入站内通知（带订单）。
func BotInsertSubscriptionOrderNotification(tx *gorm.DB, order *entity.UserOrder, sub *entity.UserSubscription, templateCode string, now time.Time) error {
	return insertSubscriptionOrderNotification(tx, order, sub, nil, templateCode, now)
}
