package mysql

import (
	"context"
	"encoding/json"
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

func (r *OrderImpl) CreateOrderWithPayment(ctx context.Context, order *entity.UserOrder, payment *entity.UserPayment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		payment.OrderID = order.ID
		return tx.Create(payment).Error
	})
}

func (r *OrderImpl) FindPaymentByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.UserPayment, error) {
	var row entity.UserPayment
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

func (r *OrderImpl) ApplyPaymentNotifySuccess(ctx context.Context, in repository.PaymentNotifyInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existed entity.UserPaymentCallback
		err := tx.Where("channel = ? AND idempotency_key = ?", in.Channel, in.IdempotencyKey).
			First(&existed).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var payment entity.UserPayment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("out_trade_no = ?", in.OutTradeNo).First(&payment).Error; err != nil {
			return err
		}
		if payment.Status == "success" {
			_ = r.insertCallback(tx, in, 1, "ignored")
			return nil
		}

		var order entity.UserOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&order, payment.OrderID).Error; err != nil {
			return err
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

		third := in.ThirdTradeNo
		if err := tx.Model(&payment).Updates(map[string]interface{}{
			"status":           "success",
			"paid_at":          now,
			"third_trade_no":   third,
			"raw_notify_json":  notifyRaw,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":     "completed",
			"paid_at":    now,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}

		if err := upsertSubscription(tx, &order, &product, now); err != nil {
			return err
		}

		if order.EnterpriseInvoice != 0 {
			inv := entity.UserInvoice{
				OrderID:     order.ID,
				UserID:      order.UserID,
				InvoiceType: "enterprise_vat",
				Title:       "企业开票信息待补充",
				AmountCents: order.TotalAmountCents,
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
	row := entity.UserPaymentCallback{
		Channel:        in.Channel,
		IdempotencyKey: in.IdempotencyKey,
		PayloadJSON:    datatypes.JSON(payload),
		SignatureOK:    sigOK,
		ProcessResult:  result,
		ProcessedAt:    time.Now(),
	}
	return tx.Create(&row).Error
}

func upsertSubscription(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, now time.Time) error {
	var sub entity.UserSubscription
	err := tx.Where("user_id = ? AND product_id = ? AND status = ?", order.UserID, order.ProductID, "active").
		First(&sub).Error

	tokenGrant := product.LimitTokens * int64(order.Quantity)
	periodEnd := addBillingPeriod(now, product.BillingPeriod)

	if errors.Is(err, gorm.ErrRecordNotFound) {
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
		return tx.Create(&row).Error
	}
	if err != nil {
		return err
	}

	var orderIDs []uint
	_ = json.Unmarshal(sub.Orders, &orderIDs)
	orderIDs = append(orderIDs, order.ID)
	ordersJSON, _ := json.Marshal(orderIDs)

	newExpires := periodEnd
	if sub.ExpiresAt.After(now) {
		newExpires = addBillingPeriod(sub.ExpiresAt, product.BillingPeriod)
	}

	return tx.Model(&sub).Updates(map[string]interface{}{
		"orders":              datatypes.JSON(ordersJSON),
		"base_limit_tokens":   tokenGrant,
		"limit_tokens":        tokenGrant,
		"used_tokens":         0,
		"expires_at":          newExpires,
		"period_start":        now,
		"period_end":          addBillingPeriod(now, product.BillingPeriod),
		"products_category_name": product.ProductsCategoryName,
		"sku_product_name":       product.SKUProductName,
		"updated_at":          now,
	}).Error
}

func addBillingPeriod(from time.Time, billingPeriod string) time.Time {
	switch billingPeriod {
	case "year":
		return from.AddDate(1, 0, 0)
	case "once":
		return from.AddDate(0, 0, 30)
	default:
		return from.AddDate(0, 1, 0)
	}
}
