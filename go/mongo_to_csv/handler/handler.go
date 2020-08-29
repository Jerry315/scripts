package handler

import (
	"dev/mongo_to_csv/common"
	"encoding/csv"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"reflect"
	"time"
)

func MongoClient(url string) (client *mgo.Session, err error) {
	client, err = mgo.Dial(url)
	return
}

func Query(client *mgo.Session, config common.Config) {
	var docs []map[string]interface{}
	var header []string
	f, err := os.Create("mawar.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	collection := client.DB("mawar").C("mawar_camera")
	sel := bson.M{}
	for _, field := range config.Fields {
		sel[field] = 1
		header = append(header, field)
	}
	w.Write(header)
	err = collection.Find(bson.M{}).Select(sel).All(&docs)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(-1)
	}

	for i, doc := range docs {
		if doc["created"] != nil {
			value := reflect.ValueOf(doc["created"])
			convertValue := value.Interface().(time.Time)
			doc["created"] = convertValue
			docs[i] = doc
		}
		if doc["updated"] != nil {
			value := reflect.ValueOf(doc["updated"])
			convertValue := value.Interface().(time.Time)
			doc["updated"] = convertValue
			docs[i] = doc
		}
		if doc["last_login"] != nil {
			value := reflect.ValueOf(doc["last_login"])
			convertValue := value.Interface().(time.Time)
			doc["last_login"] = convertValue
			docs[i] = doc
		}
		if doc["last_register"] != nil {
			value := reflect.ValueOf(doc["last_register"])
			convertValue := value.Interface().(time.Time)
			doc["last_register"] = convertValue
			docs[i] = doc
		}
		var tmp []string
		for _, field := range config.Fields {
			if doc[field] != nil {
				tmp = append(tmp, fmt.Sprintf("%v", doc[field]))
			} else {
				tmp = append(tmp, "")
			}
		}
		w.Write(tmp)
	}
}
