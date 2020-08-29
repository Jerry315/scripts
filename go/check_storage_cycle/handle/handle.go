package handle

import (
	"dev/check_storage_cycle/common"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func MongoClient(url string) (client *mgo.Session, err error) {
	//mongodb的操作入口
	client, err = mgo.Dial(url)
	return
}

func DevicesQuery(client *mgo.Session, logger *zap.Logger,cids []int64, fields []string, db, table string) (docs []common.DevicesDoc) {
	//查询通配数据库方法
	collection := client.DB(db).C(table)
	filter := bson.M{}
	if len(cids) > 0{
		filter = bson.M{"_id": bson.M{"$in": cids}}
	}
	sel := bson.M{}
	for _, field := range fields {
		sel[field] = 1
	}
	err := collection.Find(filter).Select(sel).All(&docs)
	if err != nil {
		logger.Error(fmt.Sprintf("select data from mawar db failed.  %v\n", err))
	}
	return
}

func MawarAppQuery(client *mgo.Session, logger *zap.Logger,cids []int64, fields []string, db, table string) (docs []common.MawarAppDoc) {
	//查询通配数据库方法
	collection := client.DB(db).C(table)
	filter := bson.M{}
	if len(cids) > 0{
		filter = bson.M{"_id": bson.M{"$in": cids}}
	}
	sel := bson.M{}
	for _, field := range fields {
		sel[field] = 1
	}
	err := collection.Find(filter).Select(sel).All(&docs)
	if err != nil {
		logger.Error(fmt.Sprintf("select data from mawar db failed.  %v\n", err))
	}
	return
}

func CameraQuery(client *mgo.Session, logger *zap.Logger, cids []int64, message_timestamp int64,fields []string, db, table string) (docs []common.CameraDoc) {
	collection := client.DB(db).C(table)
	limitNano := time.Now().UnixNano() / 1e6 - message_timestamp
	//默认过滤push_state=4，message_timestamp心跳时间在350s之内的cid
	filter := bson.M{"push_state":4,"message_timestamp":bson.M{"$gte": limitNano}}
	if len(cids) > 0{
		filter = bson.M{"_id": bson.M{"$in": cids},"push_state":4,"message_timestamp":bson.M{"$lte": limitNano}}
	}
	sel := bson.M{}
	for _, field := range fields {
		sel[field] = 1
	}
	err := collection.Find(filter).Select(sel).All(&docs)
	if err != nil {
		logger.Error(fmt.Sprintf("select data from camera db failed. %v\n", err))
	}
	return
}
