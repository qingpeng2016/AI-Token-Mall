package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/billing"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const autoRenewLeadDays = 7

// RunSubscriptionLifecycle 扫描 active 订阅，依次执行：自动续费、周期重置、过期处理。
func RunSubscriptionLifecycle(ctx context.Context, db *gorm.DB) error {
	now := time.Now()
	if err := runAutoRenewPhase(ctx, db, now); err != nil {
		return err
	}
	if err := runPeriodResetPhase(ctx, db, now); err != nil {
		return err
	}
	return runExpirePhase(ctx, db, now)
}

func listActiveSubscriptionIDs(ctx context.Context, db *gorm.DB) ([]uint, error) {
	var ids []uint
	err := db.WithContext(ctx).Model(&entity.UserSubscription{}).
		Where("status = ?", "active").
		Order("id ASC").
		Pluck("id", &ids).Error
	return ids, err
}

func runAutoRenewPhase(ctx context.Context, db *gorm.DB, now time.Time) error {
	ids, err := listActiveSubscriptionIDs(ctx, db)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := tryAutoRenewSubscription(ctx, db, id, now); err != nil {
			logger.ErrorZ(ctx, "subscription-auto-renew-failed", zap.Uint("subscription_id", id), zap.Error(err))
		}
	}
	return nil
}

func runPeriodResetPhase(ctx context.Context, db *gorm.DB, now time.Time) error {
	ids, err := listActiveSubscriptionIDs(ctx, db)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := tryPeriodResetSubscription(ctx, db, id, now); err != nil {
			logger.ErrorZ(ctx, "subscription-period-reset-failed", zap.Uint("subscription_id", id), zap.Error(err))
		}
	}
	return nil
}

func runExpirePhase(ctx context.Context, db *gorm.DB, now time.Time) error {
	ids, err := listActiveSubscriptionIDs(ctx, db)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := tryExpireSubscription(ctx, db, id, now); err != nil {
			logger.ErrorZ(ctx, "subscription-expire-failed", zap.Uint("subscription_id", id), zap.Error(err))
		}
	}
	return nil
}

func tryAutoRenewSubscription(ctx context.Context, db *gorm.DB, subID uint, now time.Time) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sub entity.UserSubscription
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", subID, "active").
			First(&sub).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if !now.Before(sub.ExpiresAt) {
			return nil
		}
		lead := sub.ExpiresAt.Add(-autoRenewLeadDays * 24 * time.Hour)
		if now.Before(lead) {
			return nil
		}
		if !subscriptionTimeEqual(sub.PeriodEnd, sub.ExpiresAt) {
			return nil
		}

		var product entity.Product
		if err := tx.First(&product, sub.ProductID).Error; err != nil {
			return err
		}
		price := product.Price
		balance, err := userWalletBalanceSum(tx, sub.UserID)
		if err != nil {
			return err
		}
		if balance.LessThan(price) {
			return nil
		}

		orderNo := fmt.Sprintf("RN%d%04d", now.Unix(), sub.UserID%10000)
		outTradeNo := fmt.Sprintf("BALRN%d%d", now.UnixNano()/1e6, sub.ID)
		paidAt := now
		order := entity.UserOrder{
			OrderNo:            orderNo,
			UserID:             sub.UserID,
			ProductID:          sub.ProductID,
			OrderType:          constants.OrderTypeRenewal,
			UserSubscriptionID: sub.ID,
			Quantity:           1,
			UnitPrice:          price,
			Status:             "completed",
			TotalAmount:        price,
			Currency:           product.Currency,
			PayChannel:         "balance",
			OutTradeNo:         outTradeNo,
			PaidAt:             &paidAt,
			ExpireAt:           now.Add(30 * time.Minute),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		updatedSub, err := renewSubscriptionForOrder(tx, &order, &product, now)
		if err != nil {
			return err
		}
		if err := reactivateSubscriptionAPIKeys(tx, updatedSub.ID, now); err != nil {
			return err
		}
		if err := insertSubscriptionOrderNotification(tx, &order, updatedSub, nil, "subscription_renewed", now); err != nil {
			return err
		}
		balanceAfter := balance.Sub(price)
		return insertOrderPayWalletFlowWithBalance(tx, &order, &product, "balance", balanceAfter, now)
	})
}

func tryPeriodResetSubscription(ctx context.Context, db *gorm.DB, subID uint, now time.Time) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sub entity.UserSubscription
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", subID, "active").
			First(&sub).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if !now.After(sub.PeriodEnd) {
			return nil
		}
		if !sub.ExpiresAt.After(sub.PeriodEnd) {
			return nil
		}

		var product entity.Product
		if err := tx.First(&product, sub.ProductID).Error; err != nil {
			return err
		}
		days := billing.PeriodDays(product.PeriodDays, product.BillingPeriod)
		newStart := sub.PeriodEnd
		newEnd := billing.AddPeriod(newStart, days)

		if err := tx.Model(&sub).Updates(map[string]interface{}{
			"period_start": newStart,
			"period_end":   newEnd,
			"limit_tokens": sub.BaseLimitTokens,
			"used_tokens":  0,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}

		return tx.Model(&entity.UserAPIKey{}).
			Where("user_subscription_id = ? AND status <> ?", sub.ID, "rotated").
			Updates(map[string]interface{}{
				"limit_tokens": sub.BaseLimitTokens,
				"used_tokens":  0,
				"updated_at":   now,
			}).Error
	})
}

func tryExpireSubscription(ctx context.Context, db *gorm.DB, subID uint, now time.Time) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sub entity.UserSubscription
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", subID, "active").
			First(&sub).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if !now.After(sub.ExpiresAt) {
			return nil
		}

		if err := tx.Model(&sub).Updates(map[string]interface{}{
			"status":     "expired",
			"updated_at": now,
		}).Error; err != nil {
			return err
		}

		return tx.Model(&entity.UserAPIKey{}).
			Where("user_subscription_id = ? AND status <> ?", sub.ID, "rotated").
			Updates(map[string]interface{}{
				"status":     "expired",
				"updated_at": now,
			}).Error
	})
}

func subscriptionTimeEqual(a, b time.Time) bool {
	return a.Unix() == b.Unix()
}

func insertOrderPayWalletFlowWithBalance(tx *gorm.DB, order *entity.UserOrder, product *entity.Product, channel string, balanceAfter decimal.Decimal, now time.Time) error {
	refType := "order"
	remark := fmt.Sprintf("订单 %s · %s · %s", order.OrderNo, product.CardTitle, channel)
	bal := balanceAfter
	row := entity.UserWalletFlow{
		UserID:       order.UserID,
		Type:         "pay",
		Amount:       order.TotalAmount.Neg(),
		BalanceAfter: &bal,
		Currency:     order.Currency,
		RefType:      &refType,
		RefID:        &order.ID,
		Remark:       &remark,
		CreatedAt:    now,
	}
	return tx.Create(&row).Error
}
