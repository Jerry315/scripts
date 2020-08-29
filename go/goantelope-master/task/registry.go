package task

import (
	"fmt"

	"github.com/robfig/cron"
)

// Registry 管理任务, 提供注册支持和启动等控制管理
type Registry struct {
	cron    *cron.Cron
	tasks   map[string]*Task
	options *RegistryOptions
}

// NewRegistry 创建新的任务注册管理器, 使用默认的锁资源
func NewRegistry(options *RegistryOptions) *Registry {
	reg := &Registry{
		cron:    cron.New(),
		tasks:   map[string]*Task{},
		options: options,
	}
	reg.cron.Start()
	return reg
}

// Register 注册任务
func (reg *Registry) Register(task *Task) error {
	if reg.tasks == nil {
		reg.tasks = map[string]*Task{}
	}

	if _, ok := reg.tasks[task.Name]; ok {
		return fmt.Errorf("task: task %s already exists", task.Name)
	}
	reg.tasks[task.Name] = task

	if reg.cron == nil {
		reg.cron = cron.New()
		reg.cron.Start()
	}

	err := reg.cron.AddJob(
		task.Cron, reg.wrapTaskFunc(
			task.Name, task.IntervalSeconds, task.fn))
	if err != nil {
		return fmt.Errorf("task: add task func to cron table failed %v", err)
	}
	return nil
}

// wrapTaskFunc 包装任务函数, 增加执行锁检查
func (reg *Registry) wrapTaskFunc(id string, intervalSeconds int64, fn Func) Func {
	return func() {
		if reg.options.locker != nil {
			ok := reg.options.locker.GetLock(id, intervalSeconds)
			if !ok {
				return
			}
			defer reg.options.locker.ReleaseLock(id)
		}
		fn()
	}
}

// RegistryOptions  任务注册器配置项
type RegistryOptions struct {
	locker Locker
}

// NewRegistryOptions 创建任务注册器配置选项
func NewRegistryOptions() *RegistryOptions {
	return &RegistryOptions{}
}

// SetLocker 设置任务锁
func (ro *RegistryOptions) SetLocker(locker Locker) {
	ro.locker = locker
}
