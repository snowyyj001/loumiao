package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//mongo
var MONGO_URI string = "127.0.0.1:27017"

//mysql
var (
	MYSQL_URI     string = "127.0.0.1:3306"
	MYSQL_DBNAME  string = "db"
	MYSQL_ACCOUNT string = "root"
	MYSQL_PASS    string = "123456"
)

//redis
var REDIS_URI string = "127.0.0.1:6379"

type DBNode struct {
	SqlUri    string `json:"sqluri"`
	DBName    string `json:"dbname"`
	DBAccount string `json:"dbaccount"`
	DBPass    string `json:"dbpass"`
	RedisUri  string `json:"redisuri"`
}

type ParentDBCfg struct {
	DBCfg DBNode `json:"db"`
}

var DBCfg ParentDBCfg

func init() {
	data, err := ioutil.ReadFile("config/cfg.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(data, &DBCfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(DBCfg.DBCfg.SqlUri) > 0 {
		MYSQL_URI = DBCfg.DBCfg.SqlUri
		MYSQL_DBNAME = DBCfg.DBCfg.DBName
		MYSQL_ACCOUNT = DBCfg.DBCfg.DBAccount
		MYSQL_PASS = DBCfg.DBCfg.DBPass
		REDIS_URI = DBCfg.DBCfg.RedisUri
	}
}
