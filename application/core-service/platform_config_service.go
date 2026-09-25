package coreservice

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	httpentity "github.com/gph-tech/fgmm-strategy-bitfinex/domain/http/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"gorm.io/gorm"
)

type PlatformConfigService struct {
	db       *gorm.DB
	repo     repository.LiquidationPlatformConfigRepo
	mu       sync.RWMutex
	cached   *entity.LiquidationPlatformConfig
	cachedAt time.Time
	cacheTTL time.Duration
}

func NewPlatformConfigService(db *gorm.DB, repo repository.LiquidationPlatformConfigRepo) *PlatformConfigService {
	ttl := 30 * time.Second
	if conf.GetEnv() == "local" {
		ttl = 0 // 联调改 sign_key 后立即生效，避免与 Python 脚本密钥不一致仍 200/401 混乱
	}
	return &PlatformConfigService{
		db:       db,
		repo:     repo,
		cacheTTL: ttl,
	}
}

func (s *PlatformConfigService) GetBeTrust(ctx context.Context) (*entity.LiquidationPlatformConfig, error) {
	s.mu.RLock()
	if s.cached != nil && time.Since(s.cachedAt) < s.cacheTTL {
		row := *s.cached
		s.mu.RUnlock()
		return &row, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached != nil && time.Since(s.cachedAt) < s.cacheTTL {
		row := *s.cached
		return &row, nil
	}

	row, err := s.repo.FindOne(ctx, s.db, map[string]interface{}{
		"is_enabled = ?": 1,
	}, "id ASC")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("platform config not found")
		}
		return nil, err
	}
	if strings.TrimSpace(row.SignKey) == "" {
		return nil, errors.New("platform sign_key empty")
	}
	cp := row
	s.cached = &cp
	s.cachedAt = time.Now()
	return &row, nil
}

func (s *PlatformConfigService) APISecret(ctx context.Context) (string, error) {
	row, err := s.GetBeTrust(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(row.SignKey), nil
}

func (s *PlatformConfigService) CallbackURL(ctx context.Context) (string, error) {
	row, err := s.GetBeTrust(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(row.CallbackURL), nil
}

// BinanceSpotAuth 从 liquidation_platform_config 读取 Binance 现货凭证；base_url/recv_window 仍走 yaml。
func (s *PlatformConfigService) BinanceSpotAuth(ctx context.Context) (httpentity.BinanceAuth, error) {
	row, err := s.GetBeTrust(ctx)
	if err != nil {
		return httpentity.BinanceAuth{}, err
	}
	key := strings.TrimSpace(row.BnAPIKey)
	secret := strings.TrimSpace(row.BnAPISecret)
	if key == "" || secret == "" {
		return httpentity.BinanceAuth{}, errors.New("binance bn_api_key/bn_api_secret empty")
	}
	return httpentity.BinanceAuth{
		APIKey:    key,
		APISecret: secret,
		BaseURL:   conf.GetBinanceBaseURL(),
	}, nil
}
