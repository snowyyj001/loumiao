// mysql
package mysqldb

import (
	"fmt"

	"github.com/snowyyj001/loumiao/log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"github.com/snowyyj001/loumiao/config"
)

var (
	DB *gorm.DB
)

//连接数据库,使用config-mysql参数
func Dial(tbs []interface{}) error {
	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", config.MYSQL_ACCOUNT, config.MYSQL_PASS, config.MYSQL_URI, config.MYSQL_DBNAME)
	engine, err := gorm.Open("mysql", url)
	if err != nil {
		panic(err)
	}

	engine.SingularTable(true)
	engine.DB().SetMaxIdleConns(1000)
	engine.DB().SetMaxOpenConns(3000)
	DB = engine

	create(tbs)

	log.Infof("mysql dail success: %s", config.MYSQL_DBNAME)

	return err
}

func create(tbs []interface{}) {
	for _, tb := range tbs {
		if !DB.HasTable(tb) {
			if err := DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(tb).Error; err != nil {
				panic(err)
			}
		} else {
			DB.AutoMigrate(tb)
		}
	}
}
