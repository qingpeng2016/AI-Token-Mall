package entity

import "time"

type UserCommissionPayoutConfig struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	UserID    uint      `gorm:"column:user_id;not null"`
	Channel   string    `gorm:"column:channel;size:16;not null"`
	QrURL     string    `gorm:"column:qr_url;size:512;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (UserCommissionPayoutConfig) TableName() string { return "user_commission_payout_config" }
