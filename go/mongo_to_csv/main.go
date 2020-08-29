package main

import (
	"dev/mongo_to_csv/common"
	"dev/mongo_to_csv/handler"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	config := common.GetConf()
	sel := bson.M{}
	for _, field := range config.Fields {
		sel[field] = 1
	}
	client, err := handler.MongoClient(config.Url)
	if err != nil {
		panic(err)
	}
	handler.Query(client, config)
}
