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
	UpdateOne(obj interface{}) error // 只确保更改一条(输入struct或map)
	Delete() error                   // 删除
	// 事务操作
	TLockUpdate(t Trans) error                 // 行锁
	TCount(t Trans) (int, error)               // 数目
	TAll(result interface{}, t Trans) error    // 保存搜索结果至
	TOne(result interface{}, t Trans) error    // 取一条记录
	TCreate(insert interface{}, t Trans) error // 插入一条记录
	TUpdate(obj interface{}, t Trans) error    // 更改
	TUpdateOne(obj interface{}, t Trans) error // 更改一条
	TDelete(t Trans) error                     // 删除
	// result
	GetResult() (Result, error) // 取返回结果
}
