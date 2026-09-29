package mysql

import (
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func userWalletBalanceSum(tx *gorm.DB, userID uint) (decimal.Decimal, error) {
	var sum decimal.Decimal
	err := tx.Model(&entity.UserWalletFlow{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&sum).Error
	return sum, err
}
