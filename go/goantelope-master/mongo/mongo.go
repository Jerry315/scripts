package mongo

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Source MongoDB 数据源类型
type Source struct {
	url     string
	db      string
	session *mgo.Session
}

// NewSource 创建一个新的 MongoDB 数据源
func NewSource(url, db string) (*Source, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Source{
		url:     url,
		db:      db,
		session: session,
	}, nil
}

// NewSourceWithInfo 根据 *DialInfo 创建一个 MongoDB 数据源
func NewSourceWithInfo(info *mgo.DialInfo) (*Source, error) {
	session, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, err
	}
	return &Source{
		url:     "",
		db:      info.Database,
		session: session,
	}, nil
}

// CopySession 返回一个拷贝的 session
func (source *Source) CopySession() *mgo.Session {
	return source.session.Copy()
}

// Collection 获取一个 *Collection 变量用于查询
func (source *Source) Collection(name string) *Collection {
	return &Collection{
		name:   name,
		db:     source.db,
		source: source,
	}
}

// Collection 封装 mgo.Collection
type Collection struct {
	name   string
	db     string
	source *Source
}

// Session 支持返回 mgo.Session
func (c Collection) Session() *mgo.Session {
	return c.source.CopySession()
}

// OriginCol 获取一个 *mgo.Collection 用于直接调用原始接口函数
func (c Collection) OriginCol(s *mgo.Session) *mgo.Collection {
	return s.DB(c.db).C(c.name)
}

// Insert 插入一个活多个数据库记录
func (c Collection) Insert(models ...interface{}) error {
	s := c.source.CopySession()
	defer s.Close()
	return s.DB(c.db).C(c.name).Insert(models...)
}

// Save 保存当前的模型对象数据到数据库中
func (c Collection) Save(m Model) error {
	s := c.source.CopySession()
	defer s.Close()

	m.UpdateTime(time.Now().UTC())
	update := bson.M{"$set": m}
	_, err := s.DB(c.db).C(c.name).UpsertId(m.GetID(), update)
	return err
}

// UpdateOne 修改单条数据记录
func (c Collection) UpdateOne(filter Filter, change map[string]interface{}) error {
	s := c.source.CopySession()
	defer s.Close()

	change["updated"] = time.Now().UTC()
	update := bson.M{"$set": change}
	return s.DB(c.db).C(c.name).Update(filter, update)
}

// UpdateMany 修改多条数据记录
func (c Collection) UpdateMany(filter Filter, change map[string]interface{}) error {
	s := c.source.CopySession()
	defer s.Close()

	change["updated"] = time.Now().UTC()
	update := bson.M{"$set": change}
	_, err := s.DB(c.db).C(c.name).UpdateAll(filter, update)
	return err
}

// SilenceUpdateMany 静默修改多条数据记录, 不更新 updated
func (c Collection) SilenceUpdateMany(filter Filter, change map[string]interface{}) error {
	s := c.source.CopySession()
	defer s.Close()

	update := bson.M{"$set": change}
	_, err := s.DB(c.db).C(c.name).UpdateAll(filter, update)
	return err
}

// Upsert 更新或插入新的数据
func (c Collection) Upsert(
	filter Filter, change, setOnInsert map[string]interface{},
) error {
	s := c.source.CopySession()
	defer s.Close()

	change["updated"] = time.Now().UTC()
	update := bson.M{"$set": change, "$setOnInsert": setOnInsert}
	_, err := s.DB(c.db).C(c.name).Upsert(filter, update)
	return err
}

// FindByID 根据 ID 查询
func (c Collection) FindByID(id, m interface{}, selectors ...Selector) error {
	filter := NewFilter().Equal("_id", id)
	return c.FindOne(filter, m, selectors...)
}

// FindOne 查找单个数据记录
func (c Collection) FindOne(filter Filter, m interface{}, selectors ...Selector) error {
	s := c.source.CopySession()
	defer s.Close()

	query := s.DB(c.db).C(c.name).Find(filter)
	if len(selectors) > 0 {
		query = query.Select(selectors[0])
	}
	return query.One(m)
}

// Find 查找多个符合添加的数据记录
func (c Collection) Find(filter Filter, models interface{}, selectors ...Selector) error {
	s := c.source.CopySession()
	defer s.Close()

	query := s.DB(c.db).C(c.name).Find(filter)
	if len(selectors) > 0 {
		query = query.Select(selectors[0])
	}
	return query.All(models)
}

// FindAndSort 排序查找数据记录
func (c Collection) FindAndSort(
	filter Filter, models interface{}, limit, skip int, sortFields ...string,
) error {
	s := c.source.CopySession()
	defer s.Close()

	return s.DB(c.db).C(c.name).Find(filter).
		Sort(sortFields...).Skip(skip).Limit(limit).All(models)
}

// SoftRemoveMany 软删除多条记录, 只更新 deleted 字段
func (c Collection) SoftRemoveMany(filter Filter) error {
	s := c.source.CopySession()
	defer s.Close()

	update := bson.M{"$set": bson.M{"deleted": 1, "updated": time.Now().UTC()}}
	_, err := s.DB(c.db).C(c.name).UpdateAll(filter, update)
	return err
}

// SoftRemoveOne 软删除单条记录, 只更新 deleted 字段
func (c Collection) SoftRemoveOne(filter Filter) error {
	s := c.source.CopySession()
	defer s.Close()

	update := bson.M{"$set": bson.M{"deleted": 1, "updated": time.Now().UTC()}}
	return s.DB(c.db).C(c.name).Update(filter, update)
}

// RemoveOne 根据条件删除单条记录
func (c Collection) RemoveOne(filter Filter) error {
	s := c.source.CopySession()
	defer s.Close()
	return s.DB(c.db).C(c.name).Remove(filter)
}

// RemoveMany 根据条件删除多条记录
func (c Collection) RemoveMany(filter Filter) (*mgo.ChangeInfo, error) {
	s := c.source.CopySession()
	defer s.Close()
	return s.DB(c.db).C(c.name).RemoveAll(filter)
}

// Count 查询符合条件的记录数量
func (c Collection) Count(filter Filter) (int, error) {
	s := c.source.CopySession()
	defer s.Close()
	filter = filter.NotEqual("deleted", 1)
	return s.DB(c.db).C(c.name).Find(filter).Count()
}

// EnsureIndex 创建索引
func (c Collection) EnsureIndex(index mgo.Index) error {
	s := c.source.CopySession()
	defer s.Close()
	return s.DB(c.db).C(c.name).EnsureIndex(index)
}

// Indexes 查询索引
func (c Collection) Indexes() (indexes []mgo.Index, err error) {
	s := c.source.CopySession()
	defer s.Close()
	return s.DB(c.db).C(c.name).Indexes()
}
