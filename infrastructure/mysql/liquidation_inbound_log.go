package mysql

import (
	"context"

	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/entity"
	"github.com/gph-tech/fgmm-strategy-bitfinex/domain/persistent/repository"
	"gorm.io/gorm"
)

type LiquidationInboundLogImpl struct {
	db *gorm.DB
}

func NewLiquidationInboundLogImpl(db *gorm.DB) repository.LiquidationInboundLogRepo {
	return &LiquidationInboundLogImpl{db: db}
}

func (r *LiquidationInboundLogImpl) dbOrTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *LiquidationInboundLogImpl) Create(ctx context.Context, tx *gorm.DB, row *entity.LiquidationInboundLog) error {
	return r.dbOrTx(tx).WithContext(ctx).Create(row).Error
}

func (r *LiquidationInboundLogImpl) UpdateResponse(ctx context.Context, tx *gorm.DB, id uint, httpStatus int, responseBody string) error {
	return r.dbOrTx(tx).WithContext(ctx).Model(&entity.LiquidationInboundLog{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"http_status":   httpStatus,
			"response_body": responseBody,
		}).Error
}

func (r *LiquidationInboundLogImpl) ListByLiquidationID(ctx context.Context, liquidationID string, limit int) ([]entity.LiquidationInboundLog, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []entity.LiquidationInboundLog
	err := r.db.WithContext(ctx).
		Where("liquidation_id = ?", liquidationID).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
