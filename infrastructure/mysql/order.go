package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderImpl struct {
	db *gorm.DB
}

func NewOrderImpl(db *gorm.DB) repository.OrderRepo {
	return &OrderImpl{db: db}
}

func (r *OrderImpl) CreateOrder(ctx context.Context, order *entity.UserOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderImpl) CreateRenewalOrderReplacingPending(ctx context.Context, order *entity.UserOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&entity.UserOrder{}).
			Where(
				"user_id = ? AND user_subscription_id = ? AND order_type = ? AND status = ?",
				order.UserID, order.UserSubscriptionID, "renewal", "pending_payment",
			).
			Updates(map[string]interface{}{
				"status":     "cancelled",
				"closed_at":  now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Create(order).Error
	})
}

func (r *OrderImpl) CreateUpgradeOrderReplacingPending(ctx context.Context, order *entity.UserOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&entity.UserOrder{}).
			Where(
				"user_id = ? AND user_subscription_id = ? AND order_type = ? AND status = ?",
				order.UserID, order.UserSubscriptionID, "upgrade", "pending_payment",
			).
			Updates(map[string]interface{}{
				"status":     "cancelled",
				"closed_at":  now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Create(order).Error
	})
}

func (r *OrderImpl) CreateQuotaAddonOrderReplacingPending(ctx context.Context, order *entity.UserOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&entity.UserOrder{}).
			Where(
				"user_id = ? AND user_subscription_id = ? AND order_type = ? AND status = ?",
				order.UserID, order.UserSubscriptionID, "quota_addon", "pending_payment",
			).
			Updates(map[string]interface{}{
				"status":     "cancelled",
				"closed_at":  now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Create(order).Error
	})
}

func (r *OrderImpl) FindOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.UserOrder, error) {
	var row entity.UserOrder
	err := r.db.WithContext(ctx).Where("out_trade_no = ?", outTradeNo).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OrderImpl) FindOrderByID(ctx context.Context, id uint) (*entity.UserOrder, error) {
	var row entity.UserOrder
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OrderImpl) FindActiveSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscription, error) {
	var row entity.UserSubscription
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ? AND status = ?", userID, productID, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OrderImpl) FindLatestSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscription, error) {
	var row entity.UserSubscription
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ? AND status IN ?", userID, productID, []string{"active", "expired"}).
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OrderImpl) FindSubscriptionForUser(ctx context.Context, userID, subscriptionID uint) (*entity.UserSubscription, error) {
	if subscriptionID == 0 {
		return nil, nil
	}
	var row entity.UserSubscription
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", subscriptionID, userID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OrderImpl) ApplyPaymentNotifySuccess(ctx context.Context, in repository.PaymentNotifyInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existed entity.PaymentCallback
		err := tx.Where("channel = ? AND idempotency_key = ?", in.Channel, in.IdempotencyKey).
			First(&existed).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var order entity.UserOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("out_trade_no = ?", in.OutTradeNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status == "completed" {
			_ = r.insertCallback(tx, in, 1, "ignored")
			return nil
		}
		if order.Status != "pending_payment" {
			return fmt.Errorf("order %s status %s not payable", order.OrderNo, order.Status)
		}

		var product entity.Product
		if err := tx.First(&product, order.ProductID).Error; err != nil {
			return err
		}

		now := time.Now()
		notifyRaw := datatypes.JSON(in.PayloadJSON)
		if len(in.PayloadJSON) == 0 {
			notifyRaw = datatypes.JSON([]byte("{}"))
		}

		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":          "completed",
			"paid_at":         now,
			"third_trade_no":  in.ThirdTradeNo,
			"raw_notify_json": notifyRaw,
			"updated_at":      now,
		}).Error; err != nil {
			return err
		}

		// 按 order_type 分发至 fulfill_purchase / fulfill_renewal / fulfill_upgrade / fulfill_quota_addon
		if err := fulfillPaidOrderByType(tx, &order, &product, in.Channel, now); err != nil {
			return err
		}

		if order.EnterpriseInvoice != 0 {
			inv := entity.UserInvoice{
				OrderID:     order.ID,
				UserID:      order.UserID,
				InvoiceType: "enterprise_vat",
				Title:       "企业开票信息待补充",
				Amount: order.TotalAmount,
				Status:      "pending",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := tx.Create(&inv).Error; err != nil {
				return err
			}
		}

		return r.insertCallback(tx, in, 1, "success")
	})
}

func (r *OrderImpl) insertCallback(tx *gorm.DB, in repository.PaymentNotifyInput, sigOK int, result string) error {
	payload := in.PayloadJSON
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	row := entity.PaymentCallback{
		Channel:        in.Channel,
		IdempotencyKey: in.IdempotencyKey,
		PayloadJSON:    datatypes.JSON(payload),
		SignatureOK:    sigOK,
		ProcessResult:  result,
		ProcessedAt:    time.Now(),
	}
	return tx.Create(&row).Error
}

