package coreservice

import (
	"context"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users repository.UserRepo
}

func NewUserService(users repository.UserRepo) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Register(ctx context.Context, req *request.RegisterUserReq) (*response.UserProfileResp, error) {
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	if email == "" && phone == "" {
		return nil, errorx.ErrParamsError
	}
	if email != "" {
		ex, err := s.users.FindByEmail(ctx, email)
		if err != nil {
			return nil, errorx.ErrDbError
		}
		if ex != nil {
			return nil, errorx.ErrUserExists
		}
	}
	if phone != "" {
		ex, err := s.users.FindByPhone(ctx, phone)
		if err != nil {
			return nil, errorx.ErrDbError
		}
		if ex != nil {
			return nil, errorx.ErrUserExists
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errorx.ErrUnknown
	}

	u := &entity.User{
		PasswordHash: string(hash),
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if email != "" {
		u.Email = &email
	}
	if phone != "" {
		u.Phone = &phone
	}
	if nick := strings.TrimSpace(req.Nickname); nick != "" {
		u.Nickname = &nick
	}
	if err := s.users.Create(ctx, nil, u); err != nil {
		return nil, errorx.ErrDbError
	}
	return toUserProfile(u), nil
}

func (s *UserService) Login(ctx context.Context, req *request.LoginUserReq) (*response.UserProfileResp, error) {
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	if email == "" && phone == "" {
		return nil, errorx.ErrParamsError
	}

	var u *entity.User
	var err error
	if email != "" {
		u, err = s.users.FindByEmail(ctx, email)
	} else {
		u, err = s.users.FindByPhone(ctx, phone)
	}
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if u == nil {
		return nil, errorx.ErrUserNotFound
	}
	if u.Status != "active" {
		return nil, errorx.ErrUserDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errorx.ErrWrongPassword
	}
	_ = s.users.UpdateLastLogin(ctx, u.ID)
	return toUserProfile(u), nil
}

func toUserProfile(u *entity.User) *response.UserProfileResp {
	resp := &response.UserProfileResp{
		ID:     u.ID,
		Status: u.Status,
	}
	if u.Email != nil {
		resp.Email = *u.Email
	}
	if u.Phone != nil {
		resp.Phone = *u.Phone
	}
	if u.Nickname != nil {
		resp.Nickname = *u.Nickname
	}
	return resp
}
