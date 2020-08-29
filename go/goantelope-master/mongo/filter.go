package mongo

import (
	"encoding/json"
	"log"

	"github.com/globalsign/mgo/bson"
)

// Filter  MongoDB 查询数据类型
type Filter bson.M

// NewFilter 创建一个新的查询变量, 默认 deleted 为非 1
func NewFilter() Filter {
	f := Filter{}
	return f.NotEqual("deleted", 1)
}

// NewStrictFilter 创建一个新的查询变量, 不对 deleted 字段进行预设
func NewStrictFilter() Filter {
	return Filter{}
}

// String 打印字符串
func (filter Filter) String() string {
	data, err := json.Marshal(filter)
	if err != nil {
		log.Printf("mongo: json marshal filter failed %v\n", err)
		return ""
	}
	return string(data)
}

// GreaterThan `$gt` 大于
func (filter Filter) GreaterThan(key string, value interface{}) Filter {
	if _, ok := filter[key]; ok {
		filter[key].(Filter)["$gt"] = value
	} else {
		filter[key] = Filter{"$gt": value}
	}
	return filter
}

// LessThan `$lt` 小于
func (filter Filter) LessThan(key string, value interface{}) Filter {
	if _, ok := filter[key]; ok {
		filter[key].(Filter)["$lt"] = value
	} else {
		filter[key] = Filter{"$lt": value}
	}
	return filter
}

// GreaterThanOrEqual `$gte` 大于或等于
func (filter Filter) GreaterThanOrEqual(key string, value interface{}) Filter {
	if _, ok := filter[key]; ok {
		filter[key].(Filter)["$gte"] = value
	} else {
		filter[key] = Filter{"$gte": value}
	}
	return filter
}

// LessThanOrEqual `$lte` 小于或等于
func (filter Filter) LessThanOrEqual(key string, value interface{}) Filter {
	if _, ok := filter[key]; ok {
		filter[key].(Filter)["$lte"] = value
	} else {
		filter[key] = Filter{"$lte": value}
	}
	return filter
}

// Equal 等于
func (filter Filter) Equal(key string, value interface{}) Filter {
	filter[key] = value
	return filter
}

// NotEqual `$ne` 不等于
func (filter Filter) NotEqual(key string, value interface{}) Filter {
	if _, ok := filter[key]; ok {
		filter[key].(Filter)["$ne"] = value
	} else {
		filter[key] = Filter{"$ne": value}
	}
	return filter
}

// OrWithFilters `$or` 或查询条件组
func (filter Filter) OrWithFilters(filters ...Filter) Filter {
	if _, ok := filter["$or"]; ok {
		filter["$or"] = append(filter["$or"].([]Filter), filters...)
	} else {
		filter["$or"] = filters
	}
	return filter
}

// AndWithFilters `$and` 并且查询条件组
func (filter Filter) AndWithFilters(filters ...Filter) Filter {
	filter["$and"] = filters
	return filter
}

// Regex `$regex` 正则表达式
func (filter Filter) Regex(key string, regex bson.RegEx) Filter {
	if _, ok := filter[key]; ok {
		filter[key].(Filter)["$regex"] = regex
	} else {
		filter[key] = Filter{"$regex": regex}
	}
	return filter
}

// InWithArray `$in` 值等于数组中数据
func (filter Filter) InWithArray(key string, array interface{}) Filter {
	if _, ok := filter[key]; ok {
		filter[key].(Filter)["$in"] = array
	} else {
		filter[key] = Filter{"$in": array}
	}
	return filter
}

// NotInWithArray `$nin` 值不等于数组中数据
func (filter Filter) NotInWithArray(key string, array interface{}) Filter {
	if _, ok := filter[key]; ok {
		filter[key].(Filter)["$nin"] = array
	} else {
		filter[key] = Filter{"$nin": array}
	}
	return filter
}

// Remove 移除指定的 key, 包括 `$and` `$or` 数组内的 Filter key
func (filter Filter) Remove(key string) Filter {
	delete(filter, key)
	if filters, ok := filter["$and"]; ok {
		if filtersArr, ok := filters.([]Filter); ok {

			for _, f := range filtersArr {
				delete(f, key)
			}
		}
	}
	if filters, ok := filter["$or"]; ok {
		if filtersArr, ok := filters.([]Filter); ok {

			for _, f := range filtersArr {
				delete(f, key)
			}
		}
	}
	return filter
}
