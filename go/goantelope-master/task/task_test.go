package task

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	mongoURL = os.Getenv("GOANTELOPE_TASK_MONGO_URL")
	mongoDB  = os.Getenv("GOANTELOPE_TASK_MONGO_DB")
)

func TestTask(t *testing.T) {
	assert := assert.New(t)

	locker, err := NewMongoLocker(mongoURL, mongoDB)
	assert.Nil(err)
	assert.NotNil(locker)

	options := NewRegistryOptions()
	options.SetLocker(locker)

	for i := 0; i < 3; i++ {
		registry := NewRegistry(options)
		task, err := NewTask("printFunc", "@every 5s", 5, makePrintFunc("printFunc", i))
		assert.Nil(err)
		err = registry.Register(task)
		assert.Nil(err)

		task2, err := NewTask("printFunc2", "@every 6s", 6, makePrintFunc("printFunc2", i))
		assert.Nil(err)
		err = registry.Register(task2)
		assert.Nil(err)
	}
	select {}
}

func TestLongShortTask(t *testing.T) {
	assert := assert.New(t)

	locker, err := NewMongoLocker(mongoURL, mongoDB)
	assert.Nil(err)
	assert.NotNil(locker)

	options := NewRegistryOptions()
	options.SetLocker(locker)

	registry := NewRegistry(options)
	task1, err := NewTask("longtask", "@every 5s", 5, makePrintFunc("longtask", 1))
	assert.Nil(err)
	err = registry.Register(task1)
	assert.Nil(err)

	task2, err := NewTask("shortask", "@every 3s", 3, makePrintFunc("shortask", 2))
	assert.Nil(err)
	err = registry.Register(task2)
	assert.Nil(err)

	select {}
}

func makePrintFunc(task string, id int) Func {
	return func() {
		time.Sleep(2 * time.Second)
		log.Println(task, "task id", id, time.Now())
	}
}
