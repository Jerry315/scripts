package mq

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Channel amqp.Channel 包装
type Channel struct {
	id        int
	retryWait int

	conn       *Connection
	cancelFunc context.CancelFunc

	// consumeChans 消费通道 <-chan amqp.Delivery
	consumeChans  map[string]*DeliveryChan
	errorNotifies []ErrorNotifyFunc

	Chan *amqp.Channel
}

// DeliveryChan 包装 <-chan amqp.Delivery
type DeliveryChan struct {
	C         <-chan amqp.Delivery
	failCount int
}

// NewChannel 创建一个新的 Channel
func NewChannel(id int, conn *Connection) (*Channel, error) {
	return &Channel{
		id:           id,
		conn:         conn,
		consumeChans: map[string]*DeliveryChan{},
	}, nil
}

// Open 打开一个 amqp.channel
func (channel *Channel) Open() error {
	c, err := channel.conn.amqpChannel()
	if err != nil {
		return err
	}
	err = channel.closeChan()
	if err != nil {
		log.Printf("mq: close channel error %v\n", err)
	}
	channel.Chan = c

	// 启动 close 监听
	channel.cancel()
	ctx, cancelFunc := context.WithCancel(context.Background())
	channel.cancelFunc = cancelFunc

	Errs := channel.Chan.NotifyClose(make(chan *amqp.Error))
	go channel.OnClose(ctx, Errs)
	return nil
}

// RegisterErrorNotify 注册错误回调函数
func (channel *Channel) RegisterErrorNotify(errFuncs ...ErrorNotifyFunc) {
	if channel.errorNotifies == nil {
		channel.errorNotifies = []ErrorNotifyFunc{}
	}
	channel.errorNotifies = append(channel.errorNotifies, errFuncs...)
}

// DoErrorNotify 执行错误回调通知
func (channel *Channel) DoErrorNotify(err error) {
	if channel.errorNotifies == nil {
		return
	}
	for _, f := range channel.errorNotifies {
		go f(err)
	}
}

// GetConsumeChan 获取一个消息的通道
func (channel *Channel) GetConsumeChan(queue, tag string) (*DeliveryChan, error) {
	dChan, ok := channel.consumeChans[tag]
	if !ok {
		msgs, err := channel.Chan.Consume(
			queue, tag, true, false, false, false, nil)
		if err != nil {
			return nil, err
		}
		dChan = &DeliveryChan{C: msgs}
		channel.consumeChans[tag] = dChan
	}

	// 超过两次失败, 尝试重置
	if dChan.failCount >= 2 {
		for {
			msgs, err := channel.Chan.Consume(
				queue, tag, true, false, false, false, nil)
			if err != nil {
				log.Printf(
					"mq: channel reset consume channel %v failed %v, will retry in %v seconds\n",
					channel.id, err, channel.retryWait)
				time.Sleep(time.Duration(channel.retryWait) * time.Second)
				continue
			}
			dChan.C = msgs
			dChan.failCount = 0
			break
		}
	}
	return dChan, nil
}

// OnClose 处理 channel 关闭, 尝试重置 channel
func (channel *Channel) OnClose(ctx context.Context, receiver chan *amqp.Error) {
	select {
	case Err := <-receiver:
		err := fmt.Errorf(
			"mq: channel %v error Code: %v, Reason: %v, Server: %v, Recover: %v",
			channel.id, Err.Code, Err.Reason, Err.Server, Err.Recover)
		log.Printf("%v\n", err)
		// 检查执行错误回调
		channel.DoErrorNotify(err)

		// 死循环一直尝试重置
		timer := time.NewTimer(0)
		for {
			select {
			case <-timer.C:
				err := channel.Open()
				if err != nil {
					channel.retryWait = updateRetryWait(channel.retryWait, 2)
					log.Printf(
						"mq: reset channel %v failed %v, retry in %v seconds\n",
						channel.id, err, channel.retryWait)
					timer.Reset(time.Duration(channel.retryWait) * time.Second)
					continue
				}
				// 无错误, 通知恢复
				channel.DoErrorNotify(ErrNoMore)
				return
			case <-ctx.Done():
				return
			}
		}
	case <-ctx.Done():
		return
	}
}

// Close 主动关闭 channel
func (channel *Channel) Close() error {
	channel.cancel()
	return channel.closeChan()
}

// cancel 调用 Channel.cancelFunc
func (channel *Channel) cancel() {
	if channel.cancelFunc != nil {
		channel.cancelFunc()
	}
}

// closeChan 关闭 amqp.Channel
func (channel *Channel) closeChan() error {
	if channel.Chan != nil {
		err := channel.Chan.Close()
		if err != nil {
			log.Printf("mq: close channel failed %v\n", err)
			return err
		}
	}
	return nil
}
