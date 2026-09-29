// 支付 notify 履约：ApplyPaymentNotifySuccess 完成订单落库后，仅通过 fulfillPaidOrderByType 分发。
// 四种 order_type 各自独立实现，勿在 switch 内写业务逻辑；新增类型请新增 fulfill_<type>.go 并在此注册。
package mysql

import (
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

func insertSubscriptionOrderNotification(tx *gorm.DB, order *entity.UserOrder, sub *entity.UserSubscription, apiKey *entity.UserAPIKey, templateCode string, now time.Time) error {
	subID := sub.ID
	sentAt := now
	row := entity.UserNotification{
		UserID:             order.UserID,
		UserSubscriptionID: &subID,
		Channel:            "in_app",
		TemplateCode:       templateCode,
		Status:             "sent",
		SentAt:             &sentAt,
		CreatedAt:          now,
	}
	if apiKey != nil {
		keyID := apiKey.ID
		row.APIKeyID = &keyID
	}
	return tx.Create(&row).Error
}

func reactivateSubscriptionAPIKeys(tx *gorm.DB, subscriptionID uint, now time.Time) error {
	return tx.Model(&entity.UserAPIKey{}).
		Where("user_subscription_id = ? AND status <> ?", subscriptionID, "rotated").
		Updates(map[string]interface{}{
			"status":     "active",
			"updated_at": now,
		}).Error
}

func insertOrderPayWalletFlow(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, channel string, now time.Time) error {
	refType := "order"
	remark := fmt.Sprintf("订单 %s · %s · %s", order.OrderNo, product.CardTitle, channel)
	row := entity.UserWalletFlow{
		UserID:      order.UserID,
		Type:        "pay",
		Amount: order.TotalAmount.Neg(),
		Currency:    order.Currency,
		RefType:     &refType,
		RefID:       &order.ID,
		Remark:      &remark,
		CreatedAt:   now,
	}
	return tx.Create(&row).Error
}

func fulfillPaidOrderByType(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, payChannel string, now time.Time) error {
	switch order.OrderType {
	case constants.OrderTypeRenewal:
		return fulfillRenewalOrderPaid(tx, order, product, payChannel, now)
	case constants.OrderTypeUpgrade:
		return fulfillUpgradeOrderPaid(tx, order, product, payChannel, now)
	case constants.OrderTypeQuotaAddon:
		return fulfillQuotaAddonOrderPaid(tx, order, product, payChannel, now)
	case constants.OrderTypePurchase, "":
		return fulfillPurchaseOrderPaid(tx, order, product, payChannel, now)
	default:
		return fmt.Errorf("order %s unknown order_type %q", order.OrderNo, order.OrderType)
	}
}
