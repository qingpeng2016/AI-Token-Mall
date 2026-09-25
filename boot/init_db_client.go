package boot

import (
	"github.com/gph-tech/fgmm-strategy-bitfinex/infrastructure/mysql"
	"gorm.io/gorm"
)

func NewDBClient() *gorm.DB {
	return mysql.GetDBClient()
}
