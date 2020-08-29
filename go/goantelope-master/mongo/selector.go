package mongo

import "github.com/globalsign/mgo/bson"

// Selector 控制选择数据返回字段的类型
type Selector bson.M

// NewSelector 创建一个新的选择器
func NewSelector(fields ...string) Selector {
	s := Selector{}
	for _, f := range fields {
		s[f] = true
	}
	return s
}

// WithFields 选择指定字段列表
func (s Selector) WithFields(fields ...string) Selector {
	return s.setFieldsValue(true, fields...)
}

// WithoutFields 指定字段不返回
func (s Selector) WithoutFields(fields ...string) Selector {
	return s.setFieldsValue(false, fields...)
}

// setFieldsValue 设置字段是否选择的值
func (s Selector) setFieldsValue(value bool, fields ...string) Selector {
	for _, f := range fields {
		s[f] = value
	}
	return s
}
