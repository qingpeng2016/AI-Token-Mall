package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UsersImpl struct {
	db *gorm.DB
}

func NewUsersImpl(db *gorm.DB) repository.UsersRepo {
	return &UsersImpl{db: db}
}

func (r *UsersImpl) Create(ctx context.Context, tx *gorm.DB, u *entity.Users) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Create(u).Error
}

func (r *UsersImpl) FindByEmail(ctx context.Context, email string) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) FindByPhone(ctx context.Context, phone string) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) FindByID(ctx context.Context, id uint) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) UpdateLastLogin(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.Users{}).Where("id = ?", id).Update("last_login_at", now).Error
}

func (r *UsersImpl) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Users{}).Count(&n).Error
	return n, err
}

type StatsImpl struct {
	db *gorm.DB
}

func NewStatsImpl(db *gorm.DB) repository.StatsRepo {
	return &StatsImpl{db: db}
}

func (r *StatsImpl) CountUsers(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Users{}).Count(&n).Error
	return n, err
}

func (r *StatsImpl) CountAccessLogs(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserAccessLogs{}).Count(&n).Error
	return n, err
}
