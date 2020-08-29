package task

// Locker 任务锁资源接口类型
type Locker interface {
	GetLock(id string, intervalSeconds int64) bool
	ReleaseLock(id string)
}
