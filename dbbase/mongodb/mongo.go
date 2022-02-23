// mongo
package MongoDB

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
	CONN_TIMEOUT = 10			//连接超时
	WRITE_TIMEOUT = 5			//写超时
	READ_TIMEOUT = 5			//读超时
)

var (
	mClient *mongo.Client
)

//连接数据库
func DialDefault() error {
	ctx, cancel := context.WithTimeout(context.Background(), CONN_TIMEOUT * time.Second)
	defer cancel()
	for _, cfg := range config.Cfg.SqlCfg {
		uri := fmt.Sprintf("mongodb://%s", cfg.SqlUri)
		llog.Debugf("mongo Dial: %s", uri)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			return err
		}
		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			llog.Fatalf("mongo ping error: %s", err.Error())
		} else {
			llog.Debugf("mongo connet successed!")
		}
		mClient = client
		break
	}
	return nil
}

//获取集合
func GetCollection(dbname, colname, keyname string) *mongo.Collection{
	collection := mClient.Database(dbname).Collection(colname)
	return collection
}

//添加单个document
func InsertOne(dbname, colname string, document interface{}) string {
	collection := mClient.Database(dbname).Collection(colname)
	ctx, cancel := context.WithTimeout(context.Background(),WRITE_TIMEOUT * time.Second)
	defer cancel()
	res, err := collection.InsertOne(ctx, document)
	if err != nil {
		llog.Errorf("InsertOne: %s", err.Error())
		return ""
	}
	return res.InsertedID.(primitive.ObjectID).String()
}

//添加多个document
func InsertMany(dbname, colname string, documents []interface{}) []string {
	collection := mClient.Database(dbname).Collection(colname)
	ctx, cancel := context.WithTimeout(context.Background(),WRITE_TIMEOUT * time.Second)
	defer cancel()
	res, err := collection.InsertMany(ctx, documents)
	if err != nil {
		llog.Errorf("InsertOne: %s", err.Error())
		return nil
	}
	rets := make([]string, len(res.InsertedIDs))
	for i, item := range res.InsertedIDs {
		rets[i] = item.(primitive.ObjectID).String()
	}
	return rets
}

//查询单个document
//args : filter, projection, sort
func FindOne(dbname, colname string, result interface{}, args ...bson.E) bool {
	collection := mClient.Database(dbname).Collection(colname)

	bfilter :=  bson.D{}
	bprojection :=  bson.D{}
	bsort :=  bson.D{}
	sz := len(args)
	if sz >= 1 {
		bfilter = bson.D{args[0]}
	}
	if sz >= 2 {
		bprojection = bson.D{args[1]}
	}
	if sz >= 3 {
		bsort = bson.D{args[2]}
	}
	opts := options.FindOne().SetSort(bsort).SetProjection(bprojection)
	ctx, cancel := context.WithTimeout(context.Background(), READ_TIMEOUT * time.Second)
	defer cancel()
	err := collection.FindOne(ctx, bfilter, opts).Decode(result)
	if err == mongo.ErrNoDocuments {
		return false
	} else if err != nil {
		llog.Errorf("FindOne: %s", err.Error())
		return false
	}
	return true
}

//查询多个document
//args : filter, projection, sort
func FindMany(dbname, colname string, args ...bson.E) (interface{}, error) {
	collection := mClient.Database(dbname).Collection(colname)

	bfilter :=  bson.D{}
	bprojection :=  bson.D{}
	bsort :=  bson.D{}
	sz := len(args)
	if sz >= 1 {
		bfilter = bson.D{args[0]}
	}
	if sz >= 2 {
		bprojection = bson.D{args[1]}
	}
	if sz >= 3 {
		bsort = bson.D{args[2]}
	}
	opts := options.Find().SetSort(bsort).SetProjection(bprojection)
	ctx, cancel := context.WithTimeout(context.Background(), READ_TIMEOUT * time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bfilter, opts)
	if err != nil{
		llog.Errorf("FindMany.Find: %s", err.Error())
		return nil, err
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		llog.Errorf("FindMany.cursor : %s", err.Error())
		return nil, err
	}
	return results, nil
}

//创建索引
func CreateIndex(dbname, colname, keyname string, unique bool) {
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
	}
}

//创建过期索引
func ExpireIndex(dbname, colname, keyname string, unique int) {
	collection := mClient.Database(dbname).Collection(colname)
	// create Index
	indexName, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.M{
				keyname: 1,
			},
			Options: options.Index().SetExpireAfterSeconds(int32(unique)),
		},
	)

	if err == nil {
		llog.Debugf("ExpireIndex success: %s", indexName)
	}
}