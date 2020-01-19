// mongo
package MongoDB

import (
	"buyugame/config"
	"fmt"

	"gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
)

type Mongo struct {
	dbname  string
	session *mgo.Session
	db      *mgo.Database
}

//连接数据库
//url：数据库地址 auth:数据库账号密码
//例如：mongo.Dial("127.0.0.1:27017", "root", "password")
//例如：mongo.Dial("127.0.0.1:27017")
func (self *Mongo) Dial(url string, auth ...string) error {
	var cs = url
	if len(auth) > 0 {
		cs = auth[0] + ":" + auth[1] + "@" + cs
	}
	fmt.Println("mongo dail : " + cs)
	session, err := mgo.Dial(cs) //连接服务器
	self.session = session
	if err != nil {
		defer session.Close()
		self.session.SetMode(mgo.Monotonic, true)
	}
	return err
}

//连接数据库
//url：数据库地址,,使用config-mongo默认参数
func (self *Mongo) DialDefault() error {
	session, err := mgo.Dial(config.MONGO_URI) //连接服务器
	self.session = session
	if err != nil {
		defer session.Close()
		self.session.SetMode(mgo.Monotonic, true)
	}
	return err
}

//使用name数据库
func (self *Mongo) Use(name string) {
	self.db = self.session.DB(name)
}

//查询多个文档
//col: 集合名,query: 条件参数
func (self *Mongo) Find(col string, query interface{}) *mgo.Query {
	return self.db.C(col).Find(query)
}

//查询单个文档
//col: 集合名,cmd: 条件参数
func (self *Mongo) FindOne(col string, cmd interface{}, result interface{}) (err error) {
	return self.db.C(col).Find(cmd).One(result)
}
