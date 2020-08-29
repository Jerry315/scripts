package task

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/robfig/cron"
)

// Func 任务函数类型
type Func func()

// Run 启动任务函数, 兼容 cron.Job 接口类型
func (f Func) Run() {
	f()
}

// Task 任务类型
type Task struct {
	ID              string
	Name            string
	Cron            string
	IntervalSeconds int64

	fn Func
}

// NewTask 创建任务
func NewTask(name, cronStr string, intervalSeconds int64, fn Func) (*Task, error) {
	_, err := cron.Parse(cronStr)
	if err != nil {
		return nil, fmt.Errorf("task: invalid cron %s, parse failed %v", cronStr, err)
	}
	return &Task{
		ID:              uuid.New().String(),
		Name:            name,
		Cron:            cronStr,
		IntervalSeconds: intervalSeconds,

		fn: fn,
	}, nil
}
