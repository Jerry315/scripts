package mongo

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// Model 数据模型接口类型
type Model interface {
	GetID() interface{}
	UpdateTime(time.Time)
}

// Base 模型基础类型
type Base struct {
	ID      interface{} `bson:"_id,omitempty"`
	Created time.Time   `bson:"created"`
	Updated time.Time   `bson:"updated"`
	Deleted int         `bson:"deleted"`
}

// NewBase 创建一个新的 Base 变量
func NewBase(ids ...interface{}) Base {
	now := time.Now().UTC()
	b := Base{
		Created: now,
		Updated: now,
	}

	if len(ids) > 0 {
		b.ID = ids[0]
	} else {
		// 默认使用 bson.ObjectId 作为主键
		b.ID = bson.NewObjectId()
	}
	return b
}

// GetID 返回数据模型 ID
func (base *Base) GetID() interface{} {
	return base.ID
}

// UpdateTime 更新模型记录更新时间
func (base *Base) UpdateTime(t time.Time) {
	base.Updated = t
}
