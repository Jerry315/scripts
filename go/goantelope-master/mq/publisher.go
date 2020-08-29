package mq

import (
	"github.com/streadway/amqp"
)

// Publisher 发布者
type Publisher struct {
	broker   *Broker
	channel  *Channel
	exchange string
}

// NewPublisher 创建 Publisher
func NewPublisher(broker *Broker, exchange string) (*Publisher, error) {
	channel, err := broker.Channel()
	if err != nil {
		return nil, err
	}
	err = channel.Open()
	if err != nil {
		return nil, err
	}
	return &Publisher{
		broker:   broker,
		channel:  channel,
		exchange: exchange,
	}, nil
}

// Publishing 发布的消息
type Publishing struct {
	amqp.Publishing

	// Key 消息的路由 key
	Key string
}

// Publish 发布消息
func (p *Publisher) Publish(msg Publishing) error {
	if p.channel == nil {
		channel, err := p.broker.Channel()
		if err != nil {
			return err
		}
		p.channel = channel
	}
	return p.channel.Chan.Publish(p.exchange, msg.Key, false, false, msg.Publishing)
}
