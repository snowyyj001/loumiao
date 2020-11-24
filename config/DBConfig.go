package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type DBNode struct {
	SqlUri    string `json:"sqluri"`
	DBName    string `json:"dbname"`
	DBAccount string `json:"dbaccount"`
	DBPass    string `json:"dbpass"`
	Master    int    `json:"master"`
}

type ParentDBCfg struct {
	RedisUri string   `json:"redisuri"`
	SqlCfg   []DBNode `json:"db"`
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
}
