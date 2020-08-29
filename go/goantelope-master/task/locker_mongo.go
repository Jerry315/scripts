package task

import (
	"log"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"git.topvdn.com/web/goantelope/mongo"
)

const (
	mongoLockCollectionName = "task_mongo_lock"
)

// 锁状态
var (
	MongoLockStateSet   = "set"   // 已经设置锁
	MongoLockStateUnset = "unset" // 未设置锁
)

var (
	// DefaultLockTimeoutSeconds 默认锁超时秒数, 1 小时
	DefaultLockTimeoutSeconds = 3600
)

// MongoLocker 使用 MongoDB 作为存储的 Locker
type MongoLocker struct {
	url    string
	db     string
	source *mongo.Source
}

// MongoLock 基于 MongoDB 的任务锁类型
type MongoLock struct {
	mongo.Base `bson:",inline"`

	TaskID    string `bson:"task_id"`
	Timestamp int64  `bson:"timestamp"`
	LockState string `bson:"lock_state"`
}

// NewMongoLocker 创建一个基于 MongoDB 存储的锁服务
func NewMongoLocker(url, db string) (*MongoLocker, error) {
	source, err := mongo.NewSource(url, db)
	if err != nil {
		return nil, err
	}
	return &MongoLocker{
		url:    url,
		db:     db,
		source: source,
	}, nil
}

// GetLock 获取基于 MongoDB 存储的指定任务锁
// 以下情况可以获取成功
//
// 1. 创建新的锁记录
// 2. LockState 为 `unset`
// 3. 设置时间超时: (nowTimestamp - Timestamp) > DefaultLockTimeoutSeconds
// 4. 设置时间超过 intervalSeconds: (nowTimestamp - intervalSeconds) >= timestamp
func (ml *MongoLocker) GetLock(id string, intervalSeconds int64) bool {
	log.Printf("task: mongo locker getting task %v lock.\n", id)
	ok := ml.EnsureExist(id)
	if !ok {
		return false
	}

	now := time.Now()
	nowTimestamp := now.Unix()
	timeoutTimestamp := nowTimestamp - int64(DefaultLockTimeoutSeconds)

	// id
	idFilter := mongo.NewFilter().Equal("task_id", id)
	// 状态为 `unset`
	unsetFilter := mongo.NewFilter().Equal("lock_state", MongoLockStateUnset)
	// 已经超时
	timeoutFilter := mongo.NewFilter().LessThanOrEqual("timestamp", timeoutTimestamp)
	// 间隔范围超时
	lastScheduleTimestamp := nowTimestamp - intervalSeconds
	intervalTimeoutFilter := mongo.NewFilter().LessThanOrEqual(
		"timestamp", lastScheduleTimestamp)

	orFilter := mongo.NewFilter().OrWithFilters(
		unsetFilter, intervalTimeoutFilter, timeoutFilter)
	filter := mongo.NewFilter().AndWithFilters(idFilter, orFilter)

	update := bson.M{
		"$set": bson.M{
			"timestamp":  nowTimestamp,
			"lock_state": MongoLockStateSet,
			"updated":    now,
		},
	}

	s := ml.source.CopySession()
	defer s.Close()

	ch := mgo.Change{
		Update:    update,
		Upsert:    false,
		ReturnNew: false,
	}
	changeInfo, err := s.DB(ml.db).C(mongoLockCollectionName).Find(filter).Apply(ch, nil)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Printf("task: mongo locker get %v task lock failed: %v\n", id, err)
		}
		return false
	}
	// 匹配一条记录并修改成功一条则获取成功
	if changeInfo.Matched == 1 && changeInfo.Updated == 1 {
		return true
	}
	return false
}

// ReleaseLock 释放基于 MongoDB 存储的指定任务锁
func (ml *MongoLocker) ReleaseLock(id string) {
	log.Printf("task: mongo locker releasing task %v lock.\n", id)
	s := ml.source.CopySession()
	defer s.Close()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"lock_state": MongoLockStateUnset,
			"timestamp":  now.Unix(),
			"updated":    now,
		},
	}
	filter := mongo.NewFilter().Equal("task_id", id)
	err := s.DB(ml.db).C(mongoLockCollectionName).Update(filter, update)
	if err != nil {
		log.Printf("task: mongo locker release %v task lock failed: %v.\n", id, err)
	}
}

// EnsureExist 确认存在指定锁记录
func (ml *MongoLocker) EnsureExist(id string) bool {
	s := ml.source.CopySession()
	defer s.Close()

	now := time.Now()
	setOnInsert := bson.M{
		"task_id": id, "created": now, "updated": now,
		"timestamp": 0, "lock_state": MongoLockStateUnset,
	}
	ch := mgo.Change{
		Update:    bson.M{"$setOnInsert": setOnInsert},
		Upsert:    true,
		ReturnNew: false,
	}
	filter := mongo.NewFilter().Equal("task_id", id)
	changeInfo, err := s.DB(ml.db).C(mongoLockCollectionName).Find(filter).Apply(ch, nil)
	if err != nil {
		log.Printf("task: mongo locker create %v task lock failed: %v.\n", id, err)
		return false
	}

	// 没有匹配记录但是创建失败则返回 false
	if changeInfo.Matched == 0 && changeInfo.UpsertedId == nil {
		log.Printf("task: mongo locker init insert %v task lock failed.\n", id)
		return false
	}
	return true
}
