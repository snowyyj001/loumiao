// mysql
package mysqldb

import (
	"fmt"
	"sync/atomic"

	"github.com/snowyyj001/loumiao/log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"github.com/snowyyj001/loumiao/config"
)

const (
	POOL_IDLE = 10
	POOL_MAX  = 100
)

var (
	Master *gorm.DB
	Slaves []*gorm.DB
	SLen   int32
	SIndex int32
)

//主数据库
func DBM() *gorm.DB {
	return Master
}

//从数据库
func DBS() *gorm.DB {
	if SLen == 0 {
		return Master
	}
	atomic.AddInt32(&SIndex, 1)
	atomic.CompareAndSwapInt32(&SIndex, SLen, 0)
	return Slaves[SIndex]

}

//连接数据库,使用config-mysql参数
func Dial(tbs []interface{}) error {
	for _, cfg := range config.DBCfg.SqlCfg {
		url := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", cfg.DBAccount, cfg.DBPass, cfg.SqlUri, cfg.DBName)
		engine, err := gorm.Open("mysql", url)
		if err != nil {
			panic(err)
		}
		engine.SingularTable(true)
		engine.DB().SetMaxIdleConns(POOL_IDLE)
		engine.DB().SetMaxOpenConns(POOL_MAX)
		if cfg.Master == 1 { //主数据库
			if Master != nil {
				log.Fatalf("数据库配置错误：%v", cfg)
			}
			Master = engine
		} else {
			Slaves = append(Slaves, engine)
			SLen++
		}
	}

	if tbs != nil {
		create(tbs)
	}

	log.Infof("mysql dail success: %v", config.DBCfg.SqlCfg)

	return nil
}

//连接数据库,使用指定参数
func DialDB(uri string, idle int, open int) (*gorm.DB, error) {
	engine, err := gorm.Open("mysql", uri)
	if err != nil {
		panic(err)
	}
	engine.SingularTable(true)
	engine.DB().SetMaxIdleConns(idle)
	engine.DB().SetMaxOpenConns(open)
	return engine, nil
}

//创建数据表
func create(tbs []interface{}) {
	for _, tb := range tbs {
		if !Master.HasTable(tb) {
			if err := Master.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(tb).Error; err != nil {
				panic(err)
			}
		} else {
			Master.AutoMigrate(tb)
		}
	}
}

//关闭连接
func Close() {
	if Master != nil {
		Master.Close()
		Master = nil
	}
	for _, slave := range Slaves {
		slave.Close()
	}
	Slaves = []*gorm.DB{}
	SLen = 0
	SIndex = 0
}
