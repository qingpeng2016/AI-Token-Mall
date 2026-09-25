package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserImpl struct {
	db *gorm.DB
}

func NewUserImpl(db *gorm.DB) repository.UserRepo {
	return &UserImpl{db: db}
}

func (r *UserImpl) Create(ctx context.Context, tx *gorm.DB, u *entity.User) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Create(u).Error
}

func (r *UserImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserImpl) FindByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var u entity.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserImpl) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	var u entity.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserImpl) UpdateLastLogin(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("last_login_at", now).Error
}

func (r *UserImpl) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&n).Error
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
	err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&n).Error
	return n, err
}

func (r *StatsImpl) CountAccessLogs(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserAccessLog{}).Count(&n).Error
	return n, err
}
