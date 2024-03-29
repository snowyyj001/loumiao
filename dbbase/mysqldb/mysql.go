// mysql
package mysqldb

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/snowyyj001/loumiao/lconfig"
	"github.com/snowyyj001/loumiao/lutil"
	"gorm.io/gorm/clause"
	"hash/crc32"
	"reflect"
	"sync"
	"time"

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
	SLen     int
	SIndex   int
	mLock    sync.Mutex
	logLevel = logger.Info
)

/*
NOTE TableName doesn’t allow dynamic name, its result will be cached for future
*/

/*
GORM provides First, Take, Last methods to retrieve a single object from the database,
it adds LIMIT 1 condition when querying the database,
and it will return the error ErrRecordNotFound if no record is found.
因此，如果使用这三个方法，err错误处理要注意
*/

func init() {
	engine, _ := gorm.Open(mysql.Open(""), &gorm.Config{
		DryRun: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `user`
		},
		Logger:                 newloger().LogMode(logger.Silent),
		SkipDefaultTransaction: true, //创建、更新、删除，禁用事务提交的方式
	})
	Master = engine
}

// 生成call调用后的执行的mysql语句
func Explain(call func() *gorm.DB) (string, bool) {
	tx := call()
	if tx.Error != nil {
		llog.Errorf("mysql Explain: %s", tx.Error.Error())
		return "", false
	}
	stmt := tx.Statement
	strsql := tx.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
	return strsql, true
}

type ORMDB struct {
	m_Db *gorm.DB
}

func (self *ORMDB) DB() *gorm.DB {
	return self.m_Db
}

// 主数据库
func DBM() *gorm.DB {
	return Master
}

func MTble(tname string) *gorm.DB {
	return Master.Table(tname)
}

// 从数据库
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

// 连接数据库,使用config-mysql参数,同时创建修改表
func Dial(tbs []interface{}) error {

	//account:pass@tcp(url)/dbname
	uri := fmt.Sprintf("%s?charset=utf8&parseTime=True&loc=Local", lconfig.Cfg.DBUri)
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
	Master = engine

	if tbs != nil {
		create(Master, tbs)
	}

	lutil.Go(func() { //每秒钟检测一次数据库连接状态
		for {
			sqlDB, _ := Master.DB()
			if err := sqlDB.Ping(); err != nil {
				llog.Errorf("mysql master db ping error: %s", err.Error())
				llog.Info("begin reconnect mysql")
				err = DialDefault()
				llog.Info("end reconnect mysql")
				if err == nil { //重连成功就退出，否则就不停重试
					break
				}
			}
			for i := 0; i < SLen; i++ {
				sqlDB, _ := Slaves[i].DB()
				if err := sqlDB.Ping(); err != nil {
					llog.Errorf("mysql slave[%s] db ping error: %s", lconfig.Cfg.DBUri, err.Error())
					Slaves[i] = Master
				}
			}
			time.Sleep(time.Second * 3)
		}
	})

	llog.Infof("mysql dail success: %v", lconfig.Cfg.DBUri)

	return nil
}

// 连接数据库,使用config-mysql参数,不创建修改表
func DialDefault() error {
	return Dial(nil)
}

// 创建数据表
func create(db *gorm.DB, tbs []interface{}) {
	for _, tb := range tbs {
		err := db.Table(tb.(schema.Tabler).TableName()).Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(tb)
		if err != nil {
			llog.Fatalf("create mysql table error: %s", err.Error())
		}
	}
}

// 使用主库添加st结构
func MInsert(st schema.Tabler) bool {
	err := Master.Table(st.TableName()).Create(st).Error
	if err != nil {
		llog.Errorf("MInsert: %s", err.Error())
		return false
	}
	return true
}

// 使用主库添加st结构
func MFirstOrCreate(st interface{}, conds ...interface{}) bool {
	err := Master.FirstOrCreate(st, conds...).Error
	if err != nil {
		llog.Errorf("MFirstOrCreate: %s", err.Error())
		return false
	}
	return true
}

// 使用主库更新pst结构,pst是指针，指示key，st支持 struct 和 map[string]interface{} 参数
// st为结构体时，只会更新非零值的字段
func MUpdates(pst schema.Tabler, st interface{}) bool {
	err := Master.Model(pst).Updates(st).Error
	if err != nil {
		llog.Errorf("MUpdates: %s", err.Error())
		return false
	}
	return true
}

// 使用主库更新st结构的attrs字段
func MUpdate(st schema.Tabler, attrs ...interface{}) bool {
	sz := len(attrs)
	if sz == 0 {
		err := Master.Table(st.TableName()).Select("*").UpdateColumns(st).Error
		if err != nil {
			llog.Errorf("MUpdate: %s", err.Error())
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
			llog.Errorf("MUpdate: %s", err.Error())
			return false
		}
	}
	return true
}

// 使用主库删除st结构
func MDelete(st schema.Tabler) bool {
	err := Master.Table(st.TableName()).Delete(st).Error
	if err != nil {
		llog.Errorf("MDelete: %s", err.Error())
		return false
	}
	return true
}

// 使用主库执行一个事务
func MTransaction(call func(db *gorm.DB, params ...interface{}) error, params ...interface{}) error {
	err := Master.Transaction(func(tx *gorm.DB) error {
		return call(tx, params...)
	})
	if err != nil {
		llog.Errorf("MTransaction: %s", err.Error())
		return err
	}
	return err
}

// 使用主库添加数据，冲突的话更新 ON DUPLICATE KEY UPDATE
// coloms为空时，更新除主键以外的所有列到新值
func MDuplicate(st schema.Tabler, coloms []string) bool {
	if len(coloms) == 0 {
		Master.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Table(st.TableName()).Create(st)
	} else {
		Master.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(coloms),
		}).Table(st.TableName()).Create(st)
	}
	return true
}

// 计算orm结构的CRC32值
func MGetCRCCode(st schema.Tabler) uint32 {
	v := reflect.ValueOf(st)
	tv := v.FieldByName("TFlag")
	record := tv.Int()
	tv.SetInt(0)

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf) // will read from buf
	err := encoder.Encode(st)
	if err != nil {
		llog.Errorf("MGetCRCCode： st = %v", st)
		return 0
	}
	b := buf.Bytes()
	r := crc32.ChecksumIEEE(b)

	tv.SetInt(record)

	return r
}
