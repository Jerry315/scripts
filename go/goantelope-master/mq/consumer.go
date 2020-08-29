package mq

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Consumer 消费者
type Consumer struct {
	*Dispatcher

	name  string
	queue string

	broker *Broker
	handle HandleFunc
}

// Delivery 收到的消息
type Delivery struct {
	amqp.Delivery
}

// Empty 判断收到的消息是否为空
func (d *Delivery) Empty() bool {
	return d.DeliveryTag == 0 && d.Exchange == ""
}

// NewConsumer 创建一个消费者
func NewConsumer(name, queue string, broker *Broker) *Consumer {
	return &Consumer{
		Dispatcher: NewDispatcher(0),
		name:       name,
		queue:      queue,
		broker:     broker,
	}
}

// SetHandle 设置消息处理函数
func (c *Consumer) SetHandle(handle HandleFunc) {
	if c.handle != nil {
		log.Println("mq: replacing message handle function")
	}
	c.handle = handle
}

// Active 启动消息处理 worker, 需要先设置 handle
func (c *Consumer) Active(num uint) error {
	if c.handle == nil {
		return ErrHandleNotSet
	}

	c.GoWorkers(c.handle, num)
	c.Run()

	// 启动消费
	for i := num; i > 0; i-- {
		channel, err := c.broker.Channel()
		if err != nil {
			// 停止已启动的 worker, 防止 goroutine 泄漏
			c.Stop()
			return err
		}

		err = channel.Open()
		if err != nil {
			return err
		}

		consumeTag := consumerTag(c.name, i)
		// 预创建消费通道, 如果失败则直接返回
		_, err = channel.GetConsumeChan(c.queue, consumeTag)
		if err != nil {
			return err
		}

		consumeMsg := func(ctx context.Context, jobChan chan *Job) {
			var err error
			msgChan := &DeliveryChan{}

			for {
				msgChan, err = channel.GetConsumeChan(c.queue, consumeTag)
				if err != nil {
					msgChan.failCount++
					continue
				}
				select {
				case <-ctx.Done():
					return
				case msg := <-msgChan.C:
					delivery := &Delivery{Delivery: msg}
					if !delivery.Empty() {
						jobChan <- &Job{Payload: delivery}
						continue
					}
					msgChan.failCount++
				}
			}
		}

		ctx, cancelFunc := context.WithCancel(context.Background())
		// 启动消费队列的生产者
		msgProduce := func() chan *Job {
			jobChan := make(chan *Job)
			go consumeMsg(ctx, jobChan)
			return jobChan
		}
		c.GoProducer(msgProduce, cancelFunc)
	}
	return nil
}

// Deactive 停止消费
func (c *Consumer) Deactive() {
	c.Stop()
}

// consumerTag 消费 channel 名称
func consumerTag(name string, idx uint) string {
	t := time.Now().UnixNano()
	return fmt.Sprintf("mq-consumer-%v-%v_%v", name, t, idx)
}
