package config

//mongo
const MONGO_URI string = "127.0.0.1:27017"

//mysql
var (
	MYSQL_URI     string = "127.0.0.1:3306"
	MYSQL_DBNAME  string = "buyu_db"
	MYSQL_ACCOUNT string = "root"
	MYSQL_PASS    string = "123456"
)

//redis
const REDIS_URI string = "192.168.0.139:6379"
