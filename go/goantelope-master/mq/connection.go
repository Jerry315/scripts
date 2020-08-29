package mq

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Connection amqp.Connection 包装
type Connection struct {
	id        int
	retryWait int
	lastURL   string

	connection    *amqp.Connection
	cancelFunc    context.CancelFunc
	errorNotifies []ErrorNotifyFunc
}

// Dial 创建一个 Connection
func Dial(url string) (*Connection, error) {
	conn := &Connection{
		errorNotifies: []ErrorNotifyFunc{},
	}
	err := conn.Dial(url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Dial 连接 MQ, 新的连接将替换旧的
func (conn *Connection) Dial(url string) error {
	c, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	err = conn.closeConn()
	if err != nil {
		log.Printf("mq: close connection error %v\n", err)
	}
	conn.connection = c
	conn.lastURL = url

	// 启动中断监听处理 goroutine
	conn.cancel()
	ctx, cancelFunc := context.WithCancel(context.Background())
	conn.cancelFunc = cancelFunc

	// 连接成功则尝试主动关闭连接
	Errs := conn.connection.NotifyClose(make(chan *amqp.Error))
	go conn.OnClose(ctx, Errs)
	return nil
}

// RegisterErrorNotify 注册错误回调函数
func (conn *Connection) RegisterErrorNotify(errFuncs ...ErrorNotifyFunc) {
	if conn.errorNotifies == nil {
		conn.errorNotifies = []ErrorNotifyFunc{}
	}
	conn.errorNotifies = append(conn.errorNotifies, errFuncs...)
}

// DoErrorNotify 检查并执行错误回调
func (conn *Connection) DoErrorNotify(err error) {
	if conn.errorNotifies == nil {
		return
	}
	for _, f := range conn.errorNotifies {
		go f(err)
	}
}

// OnClose 处理连接中断, 尝试重连
func (conn *Connection) OnClose(ctx context.Context, receiver chan *amqp.Error) {
	select {
	case Err := <-receiver:
		err := fmt.Errorf(
			"mq: connection error Code: %v, Reason: %v, Server: %v, Recover: %v",
			Err.Code, Err.Reason, Err.Server, Err.Recover)
		log.Printf("%v\n", err)
		// 异步执行错误回调
		conn.DoErrorNotify(err)

		// 死循环一直尝试重连
		timer := time.NewTimer(0)
		for {
			select {
			case <-timer.C:
				err := conn.Dial(conn.lastURL)
				if err != nil {
					conn.retryWait = updateRetryWait(conn.retryWait, 2)
					log.Printf(
						"mq: reconnect mq %v failed %v, retry in %v seconds\n",
						conn.lastURL, err, conn.retryWait)
					timer.Reset(time.Duration(conn.retryWait) * time.Second)
					continue
				}
				// 通知错误恢复
				conn.DoErrorNotify(ErrNoMore)
				return
			case <-ctx.Done():
				return
			}
		}
	case <-ctx.Done():
		return
	}
}

// Close 主动关闭连接
func (conn *Connection) Close() error {
	conn.cancel()
	return conn.closeConn()
}

// cancel 调用 Connection.cancelFunc
func (conn *Connection) cancel() {
	if conn.cancelFunc != nil {
		conn.cancelFunc()
	}
}

// closeConn 关闭 amqp.Connection
func (conn *Connection) closeConn() error {
	if conn.connection != nil {
		err := conn.connection.Close()
		if err != nil {
			log.Printf("mq: close amqp connection failed %v\n", err)
			return err
		}
	}
	return nil
}

// Channel 获取一个 Channel
func (conn *Connection) Channel(id int) (*Channel, error) {
	return NewChannel(id, conn)
}

// amqpChannel 获取连接上的一个 amqp.Channel
func (conn *Connection) amqpChannel() (*amqp.Channel, error) {
	if conn.connection == nil {
		return nil, ErrNoConnection
	}
	return conn.connection.Channel()
}
