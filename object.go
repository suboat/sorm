package orm

import (
	"strings"
)

// Objects 数据对象操作
type Objects interface {
	// setting
	Filter(M) Objects        // 搜索结果
	Limit(int) Objects       // 限制
	Skip(int) Objects        // 跳过
	Sort(...string) Objects  // 排序
	Group(...string) Objects // 去重
	Meta() (*Meta, error)    // 摘要信息
	// operate
	Count() (int, error)             // 数目
	All(result interface{}) error    // 保存搜索结果至
	One(result interface{}) error    // 取一条记录
	Create(insert interface{}) error // 插入一条记录
	Update(record interface{}) error // 更改(输入struct或map) *** struct为覆盖更新，map为局部更新
	UpdateOne(obj interface{}) error // 只确保更改一条(输入struct或map)
	Delete() error                   // 删除
	DeleteOne() error                // 删除一条记录
	Sum(...string) ([]int, error)    // 聚合计算
	// 事务操作
	TLockUpdate(t Trans) error                 // 行锁
	TCount(t Trans) (int, error)               // 数目
	TAll(result interface{}, t Trans) error    // 保存搜索结果至
	TOne(result interface{}, t Trans) error    // 取一条记录
	TCreate(insert interface{}, t Trans) error // 插入一条记录
	TUpdate(obj interface{}, t Trans) error    // 更改
	TUpdateOne(obj interface{}, t Trans) error // 更改一条
	TDelete(t Trans) error                     // 删除
	TDeleteOne(t Trans) error                  // 删除一条记录
	// result
	GetResult() (Result, error) // 取返回结果
	// other
	With(opt *ArgObjects) Objects // 以新的日志级别运行
}

// MetaReader Meta对象操作
type MetaReader interface {
	Meta() (*Meta, error) // 摘要信息
}

// DataReader Data对象操作
type DataReader interface {
	All(result interface{}) error // 保存搜索结果至
}

// ResultReader Data对象操作
type ResultReader interface {
	MetaReader
	DataReader
}

// M 搜索条件
type M map[string]interface{}

// Scrape 按黑名单，白名单，默认值过滤自己
func (m M) Scrape(
	whiteList map[string]interface{},
	blackList map[string]interface{},
	defaultVals map[string]interface{}) error {
	return HookParseSafe(m, whiteList, blackList, defaultVals)
}

// SQL prefix: for pg syntax
func (m M) SQL(driverName string, prefixIndex int) (sql string, vals []interface{}, err error) {
	return HookParseSQL[driverName](m, prefixIndex)
}

// Map 返回自己
func (m M) Map() (n map[string]interface{}) {
	return m
}

// IsEmpty 判断是否为空
func (m M) IsEmpty() bool {
	if m == nil {
		return true
	} else if len(m) == 0 {
		return true
	} else {
		return false
	}
}

// Set 搜索条件基本方法
func (m M) Set(k string, v interface{}) {
	m[k] = v
}

// SetNotExist 尝试添加
func (m M) SetNotExist(k string, v interface{}) {
	if !m.Hav(k) {
		m.Set(k, v)
	}
}

// SetExist 尝试更改
func (m M) SetExist(k string, v interface{}) {
	if m.Hav(k) {
		m.Set(k, v)
	}
}

// Del 删除
func (m M) Del(k string) {
	delete(m, k)
}

// Get 取值
func (m M) Get(k string) interface{} {
	return m[k]
}

// GetString 取字符串
func (m M) GetString(k string) (s string) {
	if v, ok := m[k]; ok {
		if _s, _ok := v.(string); _ok {
			s = _s
		}
	}
	return
}

// Hav 判断是否存在
func (m M) Hav(k string) (ok bool) {
	_, ok = m[k]
	return
}

// Update 更新
func (m M) Update(t *M) (err error) {
	for k, v := range *t {
		m[strings.ToLower(k)] = v
	}
	return
}

// NewM 新建并初始化
func NewM() M {
	return M(make(map[string]interface{}))
}
