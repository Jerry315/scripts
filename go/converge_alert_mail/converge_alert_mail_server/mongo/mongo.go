package mongo

import (
	"dev/converge_alert_mail/converge_alert_mail_server/common"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func MongoClient(url, db, table string) (collection *mgo.Collection, err error) {
	//mongodb的操作入口
	client, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	collection = client.DB(db).C(table)
	return
}

func MgoInsert(collection *mgo.Collection, project, module, doc interface{}, logger *zap.Logger) (err error) {
	zt := common.ZeroTime()
	filter := bson.M{"project": project, "module": module, "create_time": bson.M{"$gte": zt, "$lt": time.Now().Unix()}}
	c, err := collection.Find(filter).Count()
	if err != nil {
		return
	}
	if c > 0 {
		logger.Info(fmt.Sprintf("[MgoInsertDeviceTimeout] project: %s module: %s data has inserted", project, module))
		return nil
	}
	err = collection.Insert(doc)
	if err != nil {
		logger.Error(fmt.Sprintf("[MgoInsertDeviceTimeout] project: %s module: %s insert data failed", project, module))
	} else {
		logger.Info(fmt.Sprintf("[MgoInsertDeviceTimeout] project: %s module: %s insert data success", project, module))
	}
	return
}

func MgoQuery(collection *mgo.Collection, module string, timestamp int64, docs interface{}, logger *zap.Logger) (err error) {
	zt := common.ZeroTime()
	filter := bson.M{"module": module, "create_time": bson.M{"$gte": zt, "$lt": timestamp}}
	fileds := bson.M{"_id": 0}
	err = collection.Find(filter).Select(fileds).All(docs)
	if err != nil {
		logger.Error("[MgoQueryDeviceCycle] 获取设备异常存储记录失败")
		return
	}
	return
}
