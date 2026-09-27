package coreservice

import (
	"context"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/auth"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"golang.org/x/crypto/bcrypt"
)

var cnPhonePattern = regexp.MustCompile(`^1\d{10}$`)

type UserService struct {
	users      repository.UserRepo
	jwtSecret  string
	tokenTTL   time.Duration
}

func NewUserService(users repository.UserRepo, cfg *conf.Config) *UserService {
	secret := "ai-token-mall-dev-secret"
	ttl := 720 * time.Hour
	if cfg != nil && cfg.AuthConf != nil {
		if s := strings.TrimSpace(cfg.AuthConf.JWTSecret); s != "" {
			secret = s
		}
		if cfg.AuthConf.JWTExpireHours > 0 {
			ttl = time.Duration(cfg.AuthConf.JWTExpireHours) * time.Hour
		}
	}
	return &UserService{users: users, jwtSecret: secret, tokenTTL: ttl}
}

func (s *UserService) Register(ctx context.Context, req *request.RegisterUserReq) (*response.UserProfileResp, error) {
	if req.Password != req.ConfirmPassword {
		return nil, errorx.ErrPasswordMismatch
	}

	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	if err := validateLoginIdentity(email, phone); err != nil {
		return nil, err
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
	if err := s.users.Create(ctx, nil, u); err != nil {
		return nil, errorx.ErrDbError
	}
	return toUserProfile(u), nil
}

func (s *UserService) Login(ctx context.Context, req *request.LoginUserReq) (*response.LoginUserResp, error) {
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	if err := validateLoginIdentity(email, phone); err != nil {
		return nil, err
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

	token, err := auth.IssueUserToken(u.ID, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return nil, errorx.ErrUnknown
	}

	profile := toUserProfile(u)
	return &response.LoginUserResp{
		Token: token,
		User:  *profile,
	}, nil
}

func validateLoginIdentity(email, phone string) error {
	if email == "" && phone == "" {
		return errorx.ErrParamsError
	}
	if email != "" && phone != "" {
		return errorx.ErrParamsError
	}
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return errorx.ErrInvalidCredential
		}
		return nil
	}
	if !cnPhonePattern.MatchString(phone) {
		return errorx.ErrInvalidCredential
	}
	return nil
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
