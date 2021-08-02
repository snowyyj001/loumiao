// mysql
package mysqldb

import (
	"fmt"
	"sync"

	"github.com/snowyyj001/loumiao/util"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/llog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	POOL_IDLE = 8
	POOL_MAX  = 16 //应该小于max_connections/服务节点数（lobby+login）
)

var (
	Master   *gorm.DB
	Slaves   []*gorm.DB
	SLen     int32
	SIndex   int32
	mLock    sync.Mutex
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
	//mLock.Lock()
	//defer mLock.Unlock()
	SIndex++
	if SIndex >= SLen {
		SIndex = 0
	}
	//atomic.AddInt32(&SIndex, 1)
	//atomic.CompareAndSwapInt32(&SIndex, SLen, 0) //相等并不能保证SIndex<SLen
	return Slaves[SIndex]

}

func STble(tname string) *gorm.DB {
	return DBS().Table(tname)
}

//连接数据库,使用config-mysql参数,同时创建修改表
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
		if cfg.Master == 1 { //主数据库
			sqlDB.SetMaxIdleConns(POOL_IDLE)
			sqlDB.SetMaxOpenConns(POOL_MAX)
			if Master != nil {
				return fmt.Errorf("数据库配置错误：%v", cfg)
			}
			Master = engine
		} else { //从库最大连接数根据从库数量调整
			sqlDB.SetMaxIdleConns(util.Max(POOL_IDLE/(len(config.DBCfg.SqlCfg)-1), 1))
			sqlDB.SetMaxOpenConns(util.Max(POOL_MAX/(len(config.DBCfg.SqlCfg)-1), 1))
			Slaves = append(Slaves, engine)
			SLen++
		}
	}

	if tbs != nil {
		create(Master, tbs)
	}

	llog.Infof("mysql dail success: %v", config.DBCfg.SqlCfg)

	return nil
}

//连接数据库,使用config-mysql参数,不创建修改表
func DialDefault() error {
	return Dial(nil)
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

func DialOrm(uri string, idle int, maxconn int, tbs []interface{}) *ORMDB {
	var orm *ORMDB = new(ORMDB)
	orm.m_Db = DialDB(uri, idle, maxconn)
	if tbs != nil {
		create(orm.m_Db, tbs)
	}
	return orm
}

//创建数据表
func create(db *gorm.DB, tbs []interface{}) {
	for _, tb := range tbs {
		db.Table(tb.(schema.Tabler).TableName()).Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(tb)
	}
}

//使用主库添加st结构
func MInsert(st schema.Tabler) bool {
	err := Master.Table(st.TableName()).Create(st).Error
	if err != nil {
		llog.Warningf("MInsert: %s", err.Error())
		return false
	}
	return true
}

//使用主库更新st结构的attrs字段
func MUpdate(st schema.Tabler, attrs ...interface{}) bool {
	sz := len(attrs)
	if sz == 0 {
		err := Master.Table(st.TableName()).Select("*").UpdateColumns(st).Error
		if err != nil {
			llog.Warningf("MUpdate: %s", err.Error())
			return false
		}
	} else { //这里拆分attrs为两部分，以符合Select函数的参数要求，这很奇葩
		first := attrs[0]
		var second []interface{}
		for k, arg := range attrs {
			if k > 0 {
				second = append(second, arg)
			}
		}
		err := Master.Table(st.TableName()).Select(first, second...).Updates(st).Error
		if err != nil {
			llog.Warningf("MUpdate: %s", err.Error())
			return false
		}
	}
	return true
}

//使用主库删除st结构
func MDelete(st schema.Tabler) bool {
	err := Master.Table(st.TableName()).Delete(st).Error
	if err != nil {
		llog.Errorf("MDelete: %s", err.Error())
		return false
	}
	return true
}

//使用主库执行一个事务
func MTransaction (call func(db *gorm.DB, params ...interface{}) error, params ...interface{}) bool {
	err := Master.Transaction(func(tx *gorm.DB) error {
		return call(tx, params...)
	})
	if err != nil {
		llog.Errorf("MTransaction: %s", err.Error())
		return false
	}
	return true
}
