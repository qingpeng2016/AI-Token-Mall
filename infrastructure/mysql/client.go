package mysql

import (
	mysqlclient "github.com/gph-tech/fgmm-strategy-bitfinex/common/dederi/mysql"
	"github.com/gph-tech/fgmm-strategy-bitfinex/conf"
	"gorm.io/gorm"
	"sync"
)

var db *gorm.DB
var once sync.Once

func GetDBClient() *gorm.DB {
	once.Do(func() {
		db = mysqlclient.NewDB(conf.GetMysqlMasterConf(), conf.GetMysqlReplicaConf())
	})
	return db
}
