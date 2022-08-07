// mongo
package mongodb

import (
	"context"
	"fmt"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/llog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const (
	CONN_TIMEOUT  = 10 //连接超时
	WRITE_TIMEOUT = 5  //写超时
	READ_TIMEOUT  = 5  //读超时
)

var (
	mClient *mongo.Client
	DbName  string //loumiao_1
)

func connectDB() error {
	uri := fmt.Sprintf("mongodb://%s", config.Cfg.DBUri)
	llog.Debugf("mongo Dial: %s", uri)
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), CONN_TIMEOUT*time.Second)
	defer cancel()
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	} else {
		llog.Debugf("mongo connect successes!")
	}
	mClient = client
	return nil
}

//连接数据库
func DialDefault() error {
	err := connectDB()
	if err != nil {
		llog.Fatalf("connectDB: %s", err.Error())
	}

	go func() { //每秒钟检测一次数据库连接状态
		for {
			if mClient != nil {
				//ctx, _ := context.WithTimeout(context.Background(), CONN_TIMEOUT*time.Second)
				//llog.Debug("begin mongo ping")
				err = mClient.Ping(context.TODO(), nil)
				if err != nil {
					llog.Errorf("mongo test ping error: %s", err.Error())
					mClient.Disconnect(context.TODO())
					mClient = nil
				}
				//llog.Debug("end mongo ping")
			} else {
				err = connectDB()
				if err == nil {
					llog.Info("end reconnect mongo success")
				} else {
					llog.Errorf("end reconnect mongo failed: %s", err.Error())
				}
			}
			time.Sleep(time.Second * 3)
		}
	}()
	return nil
}

// GetCollection 获取集合
func GetCollection(dbname, colname, keyname string) *mongo.Collection {
	if mClient == nil {
		return nil
	}
	collection := mClient.Database(dbname).Collection(colname)
	return collection
}

// CountDocumentsDefault 获取集合文档数，数据量大的话不要使用
func CountDocumentsDefault(colname string, filter bson.D) (int64, error) {
	return CountDocuments(DbName, colname, filter)
}

// CountDocuments 获取集合文档数，数据量大的话不要使用
func CountDocuments(dbname, colname string, filter bson.D) (int64, error) {
	if mClient == nil {
		return 0, fmt.Errorf("CountDocumentsno db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	return collection.CountDocuments(context.TODO(), filter)
}

//InsertOneDefault 添加单个document
//colname：集合名
func InsertOneDefault(colname string, document interface{}) error {
	return InsertOne(DbName, colname, document)
}

// InsertOne 添加单个document
func InsertOne(dbname, colname string, document interface{}) error {
	if mClient == nil {
		return fmt.Errorf("InsertOne db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	//ctx, cancel := context.WithTimeout(context.Background(), WRITE_TIMEOUT*time.Second)
	//defer cancel()
	_, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		return err
	}
	//primitive.ObjectIDFromHex("5eb3d668b31de5d588f42a7a")
	//return res.InsertedID.(primitive.ObjectID).String()
	return nil
}

// InsertMany 添加多个document
func InsertMany(dbname, colname string, documents []interface{}) ([]string, error) {
	if mClient == nil {
		return nil, fmt.Errorf("InsertMany db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	//ctx, cancel := context.WithTimeout(context.Background(), WRITE_TIMEOUT*time.Second)
	//defer cancel()
	res, err := collection.InsertMany(context.TODO(), documents)
	if err != nil {
		//llog.Errorf("InsertOne: %s", err.Error())
		return nil, err
	}
	rets := make([]string, len(res.InsertedIDs))
	for i, item := range res.InsertedIDs {
		rets[i] = item.(primitive.ObjectID).String()
	}
	return rets, nil
}

// FindOneByIdDefault 查询单个document
//args : filter, projection
func FindOneByIdDefault(colname string, result interface{}, id string, args ...bson.D) error {
	if mClient == nil {
		return fmt.Errorf("FindOneByIdDefault db connected")
	}
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", _id}}
	sz := len(args)
	bprojection := bson.D{}
	if sz >= 1 {
		bprojection = args[0]
	}
	return FindOne(DbName, colname, result, filter, bprojection)
}

// FindOneDefault 查询单个document
//args : filter, projection
func FindOneDefault(colname string, result interface{}, args ...bson.D) error {
	return FindOne(DbName, colname, result, args...)
}

// FindOne 查询单个document
//colname：集合名
//result：查询结果
//args : filter, projection
func FindOne(dbname, colname string, result interface{}, args ...bson.D) error {
	if mClient == nil {
		return fmt.Errorf("FindOne db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	bfilter := bson.D{}
	bprojection := bson.D{}
	sz := len(args)
	if sz >= 1 {
		bfilter = args[0]
	}
	if sz >= 2 {
		bprojection = args[1]
	}
	opts := options.FindOne().SetProjection(bprojection)
	err := collection.FindOne(context.TODO(), bfilter, opts).Decode(result)
	//fmt.Println(err, dbname, colname, bfilter, bprojection, result)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

// FindManyDefault 查询多个document
//args : filter, projection, sort
func FindManyDefault(colname string, result interface{}, args ...bson.D) error {
	return FindMany(DbName, colname, result, args...)
}

// FindMany 查询多个document
//args : filter, projection, sort
func FindMany(dbname, colname string, result interface{}, args ...bson.D) error {
	if mClient == nil {
		return fmt.Errorf("FindMany db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	bfilter := bson.D{}
	bprojection := bson.D{}
	bsort := bson.D{}
	sz := len(args)
	if sz >= 1 {
		bfilter = args[0]
	}
	if sz >= 2 {
		bprojection = args[1]
	}
	if sz >= 3 {
		bsort = args[2]
	}
	opts := options.Find().SetSort(bsort).SetProjection(bprojection)
	cursor, err := collection.Find(context.TODO(), bfilter, opts)
	if err != nil {
		return err
	}
	if err = cursor.All(context.TODO(), result); err != nil {
		llog.Errorf("FindMany: %s", err.Error())
		return err
	}
	return nil
}

// UpdateByIDDefault  更新单个document
func UpdateByIDDefault(colname string, document interface{}, id string) error {
	return UpdateByID(DbName, colname, id, document)
}

// UpdateByID  更新单个document
func UpdateByID(dbname, colname string, id string, document interface{}) error {
	if mClient == nil {
		return fmt.Errorf("UpdateByID db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{"_id", _id}}
	update := bson.D{{"$set", document}}
	_, err := collection.UpdateByID(context.TODO(), filter, update)
	if err != nil {
		//llog.Errorf("UpdateByID: %s", err.Error())
		return err
	}
	return nil
}

// UpdateOneDefault  更新单个document
func UpdateOneDefault(colname string, document interface{}, filter bson.D) error {
	return UpdateOne(DbName, colname, document, filter)
}

// UpdateOne  更新单个document
func UpdateOne(dbname, colname string, document interface{}, filter bson.D) error {
	if mClient == nil {
		return fmt.Errorf("UpdateOne db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	opts := options.Update().SetUpsert(true)
	update := bson.D{{"$set", document}}
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		//llog.Errorf("UpdateOne: %s", err.Error())
		return err
	}
	return nil
}

// UpdateMany  更新多个document
func UpdateMany(dbname, colname string, document interface{}, args ...bson.E) error {
	if mClient == nil {
		return fmt.Errorf("UpdateMany db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	opts := options.Update().SetUpsert(true)
	filter := args
	_, err := collection.UpdateMany(context.TODO(), filter, document, opts)
	if err != nil {
		//llog.Errorf("UpdateMany: %s", err.Error())
		return err
	}
	return nil
}

// ReplaceOne  替换单个document
func ReplaceOne(dbname, colname string, document interface{}, filter bson.D) error {
	if mClient == nil {
		return fmt.Errorf("ReplaceOne db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	_, err := collection.ReplaceOne(context.TODO(), filter, document)
	if err != nil {
		//llog.Errorf("ReplaceOne: %s", err.Error())
		return err
	}
	return nil
}

func DeleteOne(dbname, colname string, filter bson.D) error {
	if mClient == nil {
		return fmt.Errorf("DeleteOne db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	_, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		//llog.Errorf("DeleteOne: %s", err.Error())
		return err
	}
	return nil
}

func DeleteOneDefault(colname string, filter bson.D) error {
	return DeleteOne(DbName, colname, filter)
}

func DeleteMany(dbname, colname string, filter bson.D) error {
	if mClient == nil {
		return fmt.Errorf("DeleteMany db connected")
	}
	collection := mClient.Database(dbname).Collection(colname)
	_, err := collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		//llog.Errorf("DeleteMany: %s", err.Error())
		return err
	}
	return nil
}

func DeleteManyDefault(colname string, filter bson.D) error {
	return DeleteMany(DbName, colname, filter)
}

// CreateIndexDefault 创建索引
func CreateIndexDefault(colname, keyname string, unique bool) {
	CreateIndex(DbName, colname, keyname, unique)
}

// CreateIndex 创建索引
func CreateIndex(dbname, colname, keyname string, unique bool) {
	if mClient == nil {
		llog.Errorf("CreateIndex no db connected: %s %s", colname, keyname)
		return
	}
	collection := mClient.Database(dbname).Collection(colname)
	// create Index
	indexName, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.M{
				keyname: 1,
			},
			Options: options.Index().SetUnique(unique),
		},
	)

	if err == nil {
		llog.Debugf("CreateIndex success: %s", indexName)
	} else {
		llog.Errorf("CreateIndex error: %s", err.Error())
	}
}

// ExpireIndexDefault 创建过期索引
func ExpireIndexDefault(colname, keyname string, sec int) {
	ExpireIndex(DbName, colname, keyname, sec)
}

// ExpireIndex 创建过期索引
func ExpireIndex(dbname, colname, keyname string, sec int) {
	if mClient == nil {
		llog.Errorf("ExpireIndex no db connected: %s %s", colname, keyname)
		return
	}
	collection := mClient.Database(dbname).Collection(colname)
	// create Index
	indexName, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.M{
				keyname: 1,
			},
			Options: options.Index().SetExpireAfterSeconds(int32(sec)),
		},
	)

	if err == nil {
		llog.Debugf("ExpireIndex success: %s", indexName)
	} else {
		llog.Errorf("ExpireIndex error: %s", err.Error())
	}
}
