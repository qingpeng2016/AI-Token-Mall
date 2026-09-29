package entity

import "time"

type UserNotifications struct {
	ID                 uint       `gorm:"primaryKey;column:id"`
	UserID             uint       `gorm:"column:user_id;not null"`
	UserSubscriptionID *uint      `gorm:"column:user_subscription_id"`
	APIKeyID           *uint      `gorm:"column:api_key_id"`
	Channel            string     `gorm:"column:channel;size:32;not null"`
	TemplateCode       string     `gorm:"column:template_code;size:64;not null"`
	Status             string     `gorm:"column:status;size:32;not null"`
	SentAt             *time.Time `gorm:"column:sent_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
}

func (UserNotifications) TableName() string { return "user_notifications" }
