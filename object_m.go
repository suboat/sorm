package orm

import (
	"fmt"
	"git.yichui.net/open/orm/log"
	"sort"
	"strings"
)

// 搜索条件
type M map[string]interface{}

// 返回自己
func (m M) Map() (n map[string]interface{}) {
	return m
}

// is empty
func (m M) IsEmpty() bool {
	if m == nil {
		return true
	} else if len(m) == 0 {
		return true
	} else {
		return false
	}
}

// 搜索条件基本方法
func (m M) Set(k string, v interface{}) {
	m[k] = v
}

// 尝试添加
func (m M) SetNotExist(k string, v interface{}) {
	if m.Hav(k) == false {
		m.Set(k, v)
	}
}

// 尝试更改
func (m M) SetExist(k string, v interface{}) {
	if m.Hav(k) == true {
		m.Set(k, v)
	}
}

func (m M) Del(k string) {
	delete(m, k)
}

func (m M) Get(k string) interface{} {
	return m[k]
}

func (m M) GetString(k string) (s string) {
	if v, ok := m[k]; ok == true {
		if _s, _ok := v.(string); _ok == true {
			s = _s
		}
	}
	return
}

func (m M) Hav(k string) (ok bool) {
	_, ok = m[k]
	return
}

func (m M) Update(t *M) (err error) {
	for k, v := range *t {
		m[k] = v
	}
	return
}

// 转为sql条件语句
// TODO: 全文搜索
func parserSql(k string, v interface{}, idx int) (sql string, val interface{}, err error) {
	// 防止sql注入, key不能含有空格
	if strings.Index(k, " ") > -1 {
		err = ErrMInvalid
		return
	}

	//if s, ok := v.(string); (ok == true && len(s) > len(TagValLike)) && (s[0:len(TagValLike)] == TagValLike) {
	if s, ok := v.(string); (ok == true && len(s) > 1) && s[0] == '%' {
		_tag, _val := IsTagVal(s)
		switch _tag {
		case TagValLike:
			sql = fmt.Sprintf("%s LIKE $%d", k, idx)
			val = "%" + _val + "%"
			break
		case TagValLt:
			sql = fmt.Sprintf("%s < $%d", k, idx)
			val = _val
			break
		case TagValLte:
			sql = fmt.Sprintf("%s <= $%d", k, idx)
			val = _val
			break
		case TagValGt:
			sql = fmt.Sprintf("%s > $%d", k, idx)
			val = _val
			break
		case TagValGte:
			sql = fmt.Sprintf("%s >= $%d", k, idx)
			val = _val
			break
		case TagValNo:
			sql = fmt.Sprintf("%s != $%d", k, idx)
			val = _val
		default:
			val = v
			break
		}
	} else {
		sql = fmt.Sprintf("%s=$%d", k, idx)
		val = v
	}

	return
}

// 解析M为sql
func (m M) Sql(driverHash Hash, args ...interface{}) (sql string, vals []interface{}, err error) {
	var (
		idx     = 0
		nameLis = []string{}
	)

	// init
	vals = []interface{}{}
	switch driverHash {
	case Postgres:
		// args: prefix index
		if len(args) > 0 {
			if _idx, ok := args[0].(int); ok {
				idx = _idx
			}
		}
		break
	}

	// 相同的map,for的顺序可能不同,会导致bug
	var (
		kIdx int = 0
		kLis     = make([]string, len(m))
	)
	for k, _ := range m {
		kLis[kIdx] = k
		kIdx += 1
	}
	sort.Strings(kLis)
	//for k, v := range m {
	for _, k := range kLis {
		v := m[k]
		// and or 查询,只支持一级嵌套
		switch k {
		case TagQueryKeyOr, TagQueryKeyAnd:
			// 解析语句
			_nameLis := []string{}
			if _lis, ok := v.([]interface{}); ok == true {
				for _, v2 := range _lis {
					if _m, ok := v2.(map[string]interface{}); ok {
						for k, v := range _m {
							idx += 1
							if _sql, _val, err2 := parserSql(k, v, idx); err2 == nil {
								_nameLis = append(_nameLis, _sql)
								vals = append(vals, _val)
							} else {
								err = err2
								return
							}
						}
					}
				}
			}
			// OR 查询
			if k == TagQueryKeyOr && len(_nameLis) > 0 {
				nameLis = append(nameLis, "("+strings.Join(_nameLis, " OR ")+")")
			} else if len(_nameLis) > 0 {
				// AND 查询
				nameLis = append(nameLis, "("+strings.Join(_nameLis, " AND ")+")")
			}
			break
		default:
			//
			idx += 1
			if _sql, _val, err2 := parserSql(k, v, idx); err2 == nil {
				nameLis = append(nameLis, "("+_sql+")")
				vals = append(vals, _val)
			} else {
				err = err2
				return
			}
			break
		}
	}
	sql = strings.Join(nameLis, " AND ")
	log.Debug("SQL: ", sql, " val:", vals)
	return
}

func NewM() (m M) {
	_m := make(map[string]interface{})
	m = M(_m)
	return
}
