package mq_test

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/streadway/amqp"

	"git.topvdn.com/web/goantelope/mq"
)

var (
	urls = []string{
		"amqp://guest:guest@127.0.0.1:5672/",
	}
	exchange   = "test"
	queue      = "test_queue_1"
	bindingKey = "test_bind_key1"
)

func ExampleConsumer() {
	broker := mq.NewBroker()
	ok, errs := broker.Connect(urls)
	if !ok {
		panic(errs)
	}

	broker.PreSetupExchange(mq.Exchange{Name: exchange, Kind: "direct", Durable: true})
	broker.PreSetupQueue(mq.Queue{Name: queue, Key: bindingKey, Exchange: exchange, Durable: true})
	err := broker.DoSetup()
	if err != nil {
		panic(err)
	}

	consumer := broker.Consumer(queue)
	msgHandle := func(job *mq.Job) {
		msg, ok := job.Payload.(*mq.Delivery)
		if !ok {
			log.Printf("job payload %v not a mq.Delivery\n", job.Payload)
		}
		_ = msg
	}
	consumer.SetHandle(msgHandle)
	err = consumer.Active(40)
	if err != nil {
		panic(err)
	}
	go func() {
		err := http.ListenAndServe(":8082", nil)
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second * 10)
	log.Println(time.Now())
	select {}
}

func ExamplePublisher() {
	broker := mq.NewBroker()
	ok, errs := broker.Connect(urls)
	if !ok {
		panic(errs)
	}

	broker.PreSetupExchange(mq.Exchange{Name: exchange, Kind: "direct", Durable: true})
	broker.PreSetupQueue(mq.Queue{Name: queue, Key: bindingKey, Exchange: exchange, Durable: true})
	err := broker.DoSetup()
	if err != nil {
		panic(err)
	}

	publisher, err := mq.NewPublisher(broker, exchange)
	if err != nil {
		panic(err)
	}

	for {
		msg := mq.Publishing{Key: bindingKey}
		msg.DeliveryMode = amqp.Persistent
		msg.Timestamp = time.Now()
		msg.ContentType = "text/plain"
		msg.Body = []byte(strings.Repeat("1", 1024))
		err := publisher.Publish(msg)
		if err != nil {
			log.Printf("publish error %v\n", err)
		}
		time.Sleep(100 * time.Millisecond * 1)
	}
}
