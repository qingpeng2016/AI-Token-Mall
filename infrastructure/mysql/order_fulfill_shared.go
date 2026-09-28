package mysql

import (
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/apikey"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

func issueAPIKeyForSubscription(tx *gorm.DB, order *entity.UserOrder, sub *entity.UserSubscription, product *entity.Product, now time.Time) (*entity.UserAPIKey, error) {
	if err := tx.Model(&entity.UserAPIKey{}).
		Where("user_subscription_id = ? AND status = ?", sub.ID, "active").
		Updates(map[string]interface{}{
			"status":     "rotated",
			"rotated_at": now,
			"updated_at": now,
		}).Error; err != nil {
		return nil, err
	}

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

func insertOrderPayWalletFlow(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, channel string, now time.Time) error {
	refType := "order"
	remark := fmt.Sprintf("订单 %s · %s · %s", order.OrderNo, product.CardTitle, channel)
	row := entity.UserWalletFlow{
		UserID:      order.UserID,
		Type:        "pay",
		AmountCents: -order.TotalAmountCents,
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
	case constants.OrderTypePurchase, "":
		return fulfillPurchaseOrderPaid(tx, order, product, payChannel, now)
	default:
		return fmt.Errorf("order %s unknown order_type %q", order.OrderNo, order.OrderType)
	}
}
