package orm

import ()

// 摘要信息
type Meta struct {
	Skip   int `json:"skip"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
	Length int `json:"length"`
}

// 数据对象操作
type Objects interface {
	Filter(M) Objects                // 搜索结果
	Count() (int, error)             // 数目
	Limit(int) Objects               // 限制
	Skip(int) Objects                // 跳过
	Sort(...string) Objects          // 排序
	Meta() (*Meta, error)            // 摘要信息
	All(result interface{}) error    // 保存搜索结果至
	One(result interface{}) error    // 取一条记录
	Create(insert interface{}) error // 插入一条记录
	Update(record interface{}) error // 更改(输入struct或map) *** struct为覆盖更新，map为局部更新
	Delete() error                   // 删除
	// 其他
	UpdateOne(obj interface{}) error // 只确保更改一条(输入struct或map)
	// 事务操作
	TCreate(insert interface{}, t Trans) error // 插入一条记录
	TDelete(t Trans) error                     // 删除
	TUpdate(obj interface{}, t Trans) error    // 更改
	TUpdateOne(obj interface{}, t Trans) error // 更改一条
}
