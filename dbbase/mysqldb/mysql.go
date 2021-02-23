// mysql
package mysqldb

import (
	"fmt"
	"sync/atomic"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/llog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	POOL_IDLE = 20
	POOL_MAX  = 100
)

var (
	Master   *gorm.DB
	Slaves   []*gorm.DB
	SLen     int32
	SIndex   int32
	logLevel logger.LogLevel = logger.Info
)

type ORMDB struct {
	m_Db *gorm.DB
}

func (self *ORMDB) DB() *gorm.DB {
	return self.m_Db
}

//主数据库
func DBM() *gorm.DB {
	return Master
}

func MTble(tname string) *gorm.DB {
	return Master.Table(tname)
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

func STble(tname string) *gorm.DB {
	return DBS().Table(tname)
}

//连接数据库,使用config-mysql参数
func Dial(tbs []interface{}) error {
	for _, cfg := range config.DBCfg.SqlCfg {
		uri := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", cfg.DBAccount, cfg.DBPass, cfg.SqlUri, cfg.DBName)
		llog.Debugf("mysql Dial: %s", uri)
		engine, err := gorm.Open(mysql.Open(uri), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `user`
			},
			Logger:                 newloger().LogMode(logLevel),
			SkipDefaultTransaction: true, //创建、更新、删除，禁用事务提交的方式
		})
		if err != nil {
			return err
		}
		sqlDB, _ := engine.DB()
		sqlDB.SetMaxIdleConns(POOL_IDLE)
		sqlDB.SetMaxOpenConns(POOL_MAX)
		if cfg.Master == 1 { //主数据库
			if Master != nil {
				return fmt.Errorf("数据库配置错误：%v", cfg)
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

	llog.Infof("mysql dail success: %v", config.DBCfg.SqlCfg)

	return nil
}

func DialDB(uri string, idle int, maxconn int) *gorm.DB {
	engine, err := gorm.Open(mysql.Open(uri), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `user`
		},
		Logger:                                   newloger().LogMode(logLevel),
		SkipDefaultTransaction:                   true, //创建、更新、删除，禁用事务提交的方式
		DisableForeignKeyConstraintWhenMigrating: true, //不自动创建外键约束
	})
	if err != nil {
		return nil
	}
	sqlDB, _ := engine.DB()
	sqlDB.SetMaxIdleConns(idle)
	sqlDB.SetMaxOpenConns(maxconn)

	return engine
}

func DialOrm(uri string, idle int, maxconn int) *ORMDB {
	var orm *ORMDB = new(ORMDB)
	orm.m_Db = DialDB(uri, idle, maxconn)
	return orm
}

//创建数据表
func create(tbs []interface{}) {
	for _, tb := range tbs {
		Master.Table(tb.(schema.Tabler).TableName()).Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(tb)
	}
}

//使用主库添加st结构
func MInsert(st schema.Tabler) *gorm.DB {
	return Master.Table(st.TableName()).Create(st)
}

//使用主库更新st结构的attrs字段
func MUpdate(st schema.Tabler, attrs ...interface{}) {
	sz := len(attrs)
	if sz == 0 {
		Master.Table(st.TableName()).Select("*").Updates(st)
	} else { //这里拆分attrs为两部分，以符合Select函数的参数要求，这很奇葩
		first := attrs[0]
		var second []interface{}
		for k, arg := range attrs {
			if k > 0 {
				second = append(second, arg)
			}
		}
		Master.Table(st.TableName()).Select(first, second...).Updates(st)
	}
}

//使用主库删除st结构
func MDelete(st schema.Tabler) {
	Master.Table(st.TableName()).Delete(st)
}
