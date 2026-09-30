package mysql

import (
	"context"
	"errors"
	"time"

	"fmt"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *UsersImpl) FindByVipDomain(ctx context.Context, vipDomain string) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).Where("vip_domain = ?", vipDomain).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) ListByParentUserID(ctx context.Context, parentUserID uint, offset, limit int) ([]entity.Users, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var rows []entity.Users
	err := r.db.WithContext(ctx).
		Where("parent_user_id = ?", parentUserID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *UsersImpl) CountByParentUserID(ctx context.Context, parentUserID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Users{}).
		Where("parent_user_id = ?", parentUserID).Count(&n).Error
	return n, err
}

func (r *UsersImpl) CountValidInvitesByParentUserID(ctx context.Context, parentUserID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Users{}).
		Where("parent_user_id = ? AND invite_valid_at IS NOT NULL", parentUserID).Count(&n).Error
	return n, err
}

func (r *UsersImpl) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id uint) (*entity.Users, error) {
	var u entity.Users
	err := repository.GormDB(ctx, r.db, tx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) ApplyWalletDelta(ctx context.Context, tx *gorm.DB, userID uint, delta decimal.Decimal) (decimal.Decimal, error) {
	u, err := r.FindByIDForUpdate(ctx, tx, userID)
	if err != nil {
		return decimal.Zero, err
	}
	if u == nil {
		return decimal.Zero, gorm.ErrRecordNotFound
	}
	newBal := u.WalletBalance.Add(delta)
	if newBal.IsNegative() {
		return decimal.Zero, fmt.Errorf("insufficient wallet balance")
	}
	if err := repository.GormDB(ctx, r.db, tx).Model(&entity.Users{}).
		Where("id = ?", userID).
		Update("wallet_balance", newBal).Error; err != nil {
		return decimal.Zero, err
	}
	return newBal, nil
}

func (r *UsersImpl) UpdateLastLogin(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.Users{}).Where("id = ?", id).Update("last_login_at", now).Error
}

func (r *UsersImpl) UpdatePassword(ctx context.Context, id uint, passwordHash, passwordPlain string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.Users{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password_hash":  passwordHash,
		"password_plain": passwordPlain,
		"updated_at":     now,
	}).Error
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
