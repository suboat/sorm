package songo

import (
	"fmt"
	"sort"
	"strings"
)

const (
	sqlSepAnd = " " + SQLValAnd + " "
	sqlSepOr  = " " + SQLValOr + " "
)

// parserSQLUnit 转为sql条件语句 TODO: 全文搜索
func parserSQLUnit(k string, v interface{}, idx int, sep string) (sql string, val interface{}, err error) {
	// valid
	if len(k) != 0 {
		k = strings.ToLower(k)
	} else {
		err = ErrSongoMapKeyInvalid
		return
	}
	// 防止sql注入, key不能含有空格,括号
	for i := 0; i < len(k); i++ {
		switch k[i] {
		case ' ', '(', ')':
			err = ErrSongoMapKeyInvalid
			return
		default:
			continue
		}
	}

	var tag string // 比较符
	val = v        // 值

	// val:尝试从key中解析比较符
	if k[len(k)-1] == TagSep {
		hIdx := -1
		for i := len(k) - 2; i > -1; i-- {
			if k[i] == TagSep {
				hIdx = i
				break
			}
		}
		if hIdx > -1 {
			if hIdx == 0 {
				err = ErrSongoMapKeyInvalid
				return
			}
			tag = k[hIdx:]
			k = k[0:hIdx]
			if len(k) == 0 {
				err = ErrSongoMapKeyInvalid
				return
			}
		}
	}

	// 解析后的key已经是字段名,防注入检测
	if SafeField(k) != k {
		err = ErrSongoMapKeyInvalid
		return
	}

	// val:尝试从val中解析比较符
	if len(tag) == 0 {
		switch _val := v.(type) {
		case string:
			if _t, _v := isTagVal(_val); len(_t) > 0 {
				tag = _t
				val = _v
			}
			// break
		case map[string]interface{}:
			if len(_val) != 1 {
				// map表示的val只有一对key-val
				err = ErrSongoMapValMapMultiple
				return
			}
			for _t, _v := range _val {
				tag = _t
				val = _v
			}
			// break
		default:
			break
		}
	}

	// sql
	if len(tag) > 0 {
		switch tag {
		case TagValLike, TagValText: // TODO: 全文搜索
			if sep == "?" {
				sql = fmt.Sprintf("`%s` %s ?", k, SQLValLike)
			} else {
				sql = fmt.Sprintf(`"%s" %s $%d`, k, SQLValLike, idx)
			}
			// break
		case TagValLt:
			if sep == "?" {
				sql = fmt.Sprintf("`%s` %s ?", k, SQLValLt)
			} else {
				sql = fmt.Sprintf(`"%s" %s $%d`, k, SQLValLt, idx)
			}
			// break
		case TagValLte:
			if sep == "?" {
				sql = fmt.Sprintf("`%s` %s ?", k, SQLValLte)
			} else {
				sql = fmt.Sprintf(`"%s" %s $%d`, k, SQLValLte, idx)
			}
			// break
		case TagValGt:
			if sep == "?" {
				sql = fmt.Sprintf("`%s` %s ?", k, SQLValGt)
			} else {
				sql = fmt.Sprintf(`"%s" %s $%d`, k, SQLValGt, idx)
			}
			// break
		case TagValGte:
			if sep == "?" {
				sql = fmt.Sprintf("`%s` %s ?", k, SQLValGte)
			} else {
				sql = fmt.Sprintf(`"%s" %s $%d`, k, SQLValGte, idx)
			}
			// break
		case TagValNo, TagValNe:
			if sep == "?" {
				sql = fmt.Sprintf("`%s` %s ?", k, SQLValNe)
			} else {
				sql = fmt.Sprintf(`"%s" %s $%d`, k, SQLValNe, idx)
			}
		default:
			err = ErrSongoMapOperatorInvalid
			return
		}
	} else {
		if sep == "?" {
			sql = fmt.Sprintf("`%s` %s ?", k, SQLValEq)
		} else {
			if val == nil {
				sql = fmt.Sprintf(`"%s" IS NULL`, k)
			} else {
				sql = fmt.Sprintf(`"%s" %s $%d`, k, SQLValEq, idx)
			}
		}
	}

	return
}

// parserSQL 解析M为sql
func parserSQL(m map[string]interface{}, prefix int, sep string) (sql string, vals []interface{}, err error) {
	var (
		nameLis = []string{}
	)
	vals = []interface{}{}

	// 相同的map,for的顺序可能不同,会导致bug
	var (
		idx = 0
		lis = make([]string, len(m))
	)
	for k := range m {
		lis[idx] = k
		idx++
	}
	sort.Strings(lis)
	//for k, v := range m {
	for _, k := range lis {
		// and or 查询, 支持二级嵌套
		v := m[k]

		// 递归解析
		if err = parserSQLOper(0, &prefix, sep, k, v, &nameLis, &vals); err != nil {
			return
		}
	}
	sql = strings.Join(nameLis, fmt.Sprintf(" %s ", SQLValAnd))
	return
}

// parserSQLOper 解析操作符
func parserSQLOper(deep int, prefix *int, sep string, k string, v interface{}, nameLis *[]string, vals *[]interface{}) (err error) {
	if deep+1 >= ParseMapMax {
		err = ErrSongoMapDeepOutOf
		return
	}
	oper := ""
	oper, k = isKeyOper(k)
	switch oper {
	case TagQueryKeyOr, TagQueryKeyAnd:
		// 解析语句
		_nameLis := []string{}
		_lis, ok := v.([]interface{})
		if !ok {
			_lis = []interface{}{v} // 默认都是数组
		}
		for _, v2 := range _lis {
			switch _v := v2.(type) {
			case map[string]interface{}:
				// map

				// sort
				var (
					idx = 0
					lis = make([]string, len(_v))
				)
				for k := range _v {
					lis[idx] = k
					idx++
				}
				sort.Strings(lis)

				for _, k := range lis {
					v := _v[k]

					org := k
					oper := ""
					oper, k = isKeyOper(k)
					switch oper {
					case TagQueryKeyOr, TagQueryKeyAnd:
						// 解析语句
						_nameLis2 := []string{}
						_lis2, ok := v.([]interface{})
						if !ok {
							_lis2 = []interface{}{v} // 默认都是数组
						}

						//
						for _, v3 := range _lis2 {
							if err = parserSQLOper(deep+1, prefix, sep, org, v3, &_nameLis2, vals); err != nil {
								return
							}
						}

						// OR 查询
						if oper == TagQueryKeyOr && len(_nameLis2) > 0 {
							_nameLis = append(_nameLis, "("+strings.Join(_nameLis2, sqlSepOr)+")")
						} else if len(_nameLis2) > 0 {
							_nameLis = append(_nameLis, "("+strings.Join(_nameLis2, sqlSepAnd)+")")
						}

						// break
					default:
						*prefix++
						if _sql, _val, err2 := parserSQLUnit(k, v, *prefix, sep); err2 == nil {
							_nameLis = append(_nameLis, _sql)
							*vals = append(*vals, _val)
						} else {
							err = err2
							return
						}
						// break
					}
				}
				// break
			default:
				// string int 等
				*prefix++
				if _sql, _val, err2 := parserSQLUnit(k, _v, *prefix, sep); err2 == nil {
					_nameLis = append(_nameLis, _sql)
					*vals = append(*vals, _val)
				} else {
					err = err2
					return
				}
				// break
			}
		}

		// OR 查询
		if oper == TagQueryKeyOr && len(_nameLis) > 0 {
			*nameLis = append(*nameLis, "("+strings.Join(_nameLis, sqlSepOr)+")")
		} else if len(_nameLis) > 0 {
			// AND 查询
			*nameLis = append(*nameLis, "("+strings.Join(_nameLis, sqlSepAnd)+")")
		}

		// break
	default:
		//
		*prefix++
		if _sql, _val, err2 := parserSQLUnit(k, v, *prefix, sep); err2 == nil {
			*nameLis = append(*nameLis, "("+_sql+")")
			if _val != nil {
				*vals = append(*vals, _val)
			}
		} else {
			err = err2
			return
		}
		// break
	}

	return
}
