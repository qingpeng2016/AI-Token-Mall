package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type SubscriptionImpl struct {
	db *gorm.DB
}

func NewSubscriptionImpl(db *gorm.DB) repository.SubscriptionRepo {
	return &SubscriptionImpl{db: db}
}

func (r *SubscriptionImpl) ListByUserID(ctx context.Context, userID uint) ([]repository.UserSubscriptionListRow, error) {
	var rows []repository.UserSubscriptionListRow
	err := r.db.WithContext(ctx).
		Table("user_subscriptions AS s").
		Select("s.*, p.card_title AS product_card_title").
		Joins("LEFT JOIN products p ON p.id = s.product_id").
		Where("s.user_id = ?", userID).
		Order("s.id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
