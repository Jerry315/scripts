package handle

import (
	"dev/storage_cycle_log/common"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func MongoClient(url string) (client *mgo.Session, err error) {
	//mongodb的操作入口
	client, err = mgo.Dial(url)
	return
}

func MawarQuery(client *mgo.Session, logger *zap.Logger, fields []string, db, table string) (docs []common.MawarDoc) {
	//查询通配数据库方法
	collection := client.DB(db).C(table)
	sel := bson.M{}
	for _, field := range fields {
		sel[field] = 1
	}
	err := collection.Find(bson.M{"is_bind": true}).Select(sel).All(&docs)
	if err != nil {
		logger.Error(fmt.Sprintf("select data from mawar db failed.  %v\n", err))
	}
	return
}
