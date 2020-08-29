package mq

import (
	"errors"
)

// Error
var (
	ErrNoConfig     = errors.New("mq: broker has no config")
	ErrNoConnection = errors.New("mq: no amqp connection")

	ErrHandleNotSet = errors.New("mq: consumer message handle function not set")

	ErrNoMore = errors.New("mq: no more errors")
)

// ErrorNotifyFunc 错误通知函数类型
type ErrorNotifyFunc func(error)
