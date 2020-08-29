package mongo_test

import (
	"fmt"

	"github.com/globalsign/mgo/bson"

	"git.topvdn.com/web/goantelope/mongo"
)

// 定义常量
const (
	MongoURL        = "mongodb://127.0.0.1:27017/test"
	MongoDB         = "test"
	MongoCollection = "person"
)

// Person 人
type Person struct {
	mongo.Base `bson:",inline"`

	Name string `bson:"name"`
	Age  int    `bson:"age"`
}

// NewPerson 创建一个 Person 变量
func NewPerson(name string, age int) *Person {
	return &Person{
		Base: mongo.NewBase(),
		Name: name,
		Age:  age,
	}
}

// Example 示例代码
func Example() {
	source, err := mongo.NewSource(MongoURL, MongoDB)
	if err != nil {
		panic(err)
	}

	personCol := source.Collection(MongoCollection)

	// insert
	person := NewPerson("tony", 32)
	err = personCol.Save(person)
	if err != nil {
		fmt.Println(err)
	}

	newPerson := Person{}
	// find NewFilter 预设 deleted 字段 {$ne: 1}
	filter := mongo.NewFilter().Equal("name", "tony")
	fmt.Printf("filter: %s\n", filter)
	err = personCol.FindOne(filter, &newPerson)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(newPerson)

	// update NewStrictFilter 不预设 deleted 字段
	updates := bson.M{"age": person.Age + 3}
	updateFilter := mongo.NewStrictFilter().Equal("_id", person.ID)
	fmt.Printf("update filter: %s\n", updateFilter)
	err = personCol.UpdateOne(updateFilter, updates)
	if err != nil {
		fmt.Println(err)
	}

	// delete NewStrictFilter 不预设 deleted 字段
	deleteFilter := mongo.NewStrictFilter().Equal("_id", person.ID)
	fmt.Printf("delete filter: %s\n", deleteFilter)
	err = personCol.RemoveOne(deleteFilter)
	if err != nil {
		fmt.Println(err)
	}

	// 单字段多条件
	testFilter := mongo.NewFilter()
	testFilter.GreaterThan("n", 10)
	testFilter.LessThanOrEqual("n", 42)
	fmt.Println("test filter", testFilter.String())
}
