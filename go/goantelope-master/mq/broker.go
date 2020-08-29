package mq

import (
	"log"
)

// Broker MQ 连接中介, 支持连接 MQ 集群多个实例
type Broker struct {
	conns []*Connection

	preSetQueues    []Queue
	preSetExchanges []Exchange

	channelOpened int
}

// Exchange 交换机
type Exchange struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
}

// Queue 队列
type Queue struct {
	Name       string
	Key        string
	Exchange   string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}

// NewBroker 创建 Broker
func NewBroker() *Broker {
	return &Broker{
		preSetQueues:    []Queue{},
		preSetExchanges: []Exchange{},
	}
}

// Connect 连接 MQ, 只要有一个能连上就不会全部失败
func (broker *Broker) Connect(urls []string, errorNotifies ...ErrorNotifyFunc) (bool, []error) {
	errs := []error{}
	conns := []*Connection{}
	for _, url := range urls {
		conn, err := Dial(url)
		if err != nil {
			errs = append(errs, err)
			log.Printf("mq: connect mq %v failed %v", url, err)
			continue
		}
		conn.RegisterErrorNotify(errorNotifies...)
		conns = append(conns, conn)
	}

	// 没有成功建立新的连接直接返回, 只要一个连接成功就返回 true
	if len(conns) == 0 {
		return false, errs
	}
	broker.conns = conns
	return true, errs
}

// Channel 获取一个 MQ 的通道
func (broker *Broker) Channel(errorNotifies ...ErrorNotifyFunc) (*Channel, error) {
	idx := broker.channelOpened % len(broker.conns)
	broker.channelOpened++
	channel, err := broker.conns[idx].Channel(broker.channelOpened)
	if err != nil {
		return nil, err
	}
	channel.RegisterErrorNotify(errorNotifies...)
	return channel, nil
}

// Consumer 从 Broker 获取一个 Consumer
func (broker *Broker) Consumer(queue string) *Consumer {
	return NewConsumer(queue, queue, broker)
}

// Publisher 从 Broker 获取一个 Publisher, concurrency 控制并发的 channel 数量
func (broker *Broker) Publisher(exchange string) (*Publisher, error) {
	return NewPublisher(broker, exchange)
}

// SetupExchange 在 MQ 上创建 exchange
func (broker *Broker) SetupExchange(exchange Exchange) error {
	channel, err := broker.Channel()
	if err != nil {
		return err
	}
	err = channel.Open()
	if err != nil {
		return err
	}
	defer func() {
		err := channel.Close()
		if err != nil {
			log.Printf("mq: close channel failed %v\n", err)
		}
	}()

	return declareExchange(channel, exchange)
}

// SetupQueue 在 MQ 上创建 queue
func (broker *Broker) SetupQueue(queue Queue) error {
	channel, err := broker.Channel()
	if err != nil {
		return err
	}
	err = channel.Open()
	if err != nil {
		return err
	}
	defer func() {
		err := channel.Close()
		if err != nil {
			log.Printf("mq: close channel failed %v\n", err)
		}
	}()

	return declareAndBindQueue(channel, queue)
}

// PreSetupExchange 预创建 exchange
func (broker *Broker) PreSetupExchange(exchange Exchange) {
	if broker.preSetExchanges == nil {
		broker.preSetExchanges = []Exchange{}
	}
	broker.preSetExchanges = append(broker.preSetExchanges, exchange)
}

// PreSetupQueue 预创建 queue 并绑定 key
func (broker *Broker) PreSetupQueue(queue Queue) {
	if broker.preSetQueues == nil {
		broker.preSetQueues = []Queue{}
	}
	broker.preSetQueues = append(broker.preSetQueues, queue)
}

// DoSetup 统一进行配置 PreSetup 的 exchange 和 queue
func (broker *Broker) DoSetup() error {
	channel, err := broker.Channel()
	if err != nil {
		return err
	}
	err = channel.Open()
	if err != nil {
		return err
	}
	defer func() {
		err := channel.Close()
		if err != nil {
			log.Printf("mq: close channel failed %v\n", err)
		}
	}()

	for _, exchange := range broker.preSetExchanges {
		err := declareExchange(channel, exchange)
		if err != nil {
			return err
		}
	}
	for _, queue := range broker.preSetQueues {
		err := declareAndBindQueue(channel, queue)
		if err != nil {
			return err
		}
	}

	// 已创建则清除旧的
	broker.preSetExchanges = []Exchange{}
	broker.preSetQueues = []Queue{}
	return nil
}

func declareExchange(channel *Channel, exchange Exchange) error {
	return channel.Chan.ExchangeDeclare(
		exchange.Name, exchange.Kind, exchange.Durable,
		exchange.AutoDelete, exchange.Internal, exchange.NoWait, nil)
}

func declareAndBindQueue(channel *Channel, queue Queue) error {
	_, err := channel.Chan.QueueDeclare(
		queue.Name, queue.Durable, queue.AutoDelete,
		queue.Exclusive, queue.NoWait, nil)
	if err != nil {
		return err
	}

	// 绑定队列 Key 和 exchange
	if queue.Key != "" && queue.Exchange != "" {
		err := channel.Chan.QueueBind(
			queue.Name, queue.Key, queue.Exchange, queue.NoWait, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateRetryWait 更新重试等待时间
func updateRetryWait(wait, step int) int {
	max := 16
	// 从 1 秒开始
	if wait <= 0 {
		wait = 1
	}

	wait += step
	// 最大重连等待为 16
	if wait > max {
		wait = max
	}
	return wait
}
