package songo

import (
	//"gopkg.in/mgo.v2/bson"
	"github.com/globalsign/mgo/bson"

	"sort"
	"strings"
)

// songoTagMongo 将songo的tag转为mongo的tag
func songoTagMongo(tag string) string {
	switch tag {
	case TagValNo, TagValNe:
		tag = "$ne"
	case TagValLt:
		tag = "$lt"
	case TagValLte:
		tag = "$lte"
	case TagValGt:
		tag = "$gt"
	case TagValGte:
		tag = "$gte"
	case TagValLike:
		tag = "$regex"
	case TagQueryKeyOr:
		tag = "$or"
	case TagQueryKeyAnd:
		tag = "$and"
	case TagQueryKeyIn:
		tag = "$in"
	case TagUpdateInc:
		tag = "$inc"
	default:
		break
	}
	return tag
}

// parseMgo 解析map为mgo
func parseMgo(m map[string]interface{}) (d map[string]interface{}, err error) {
	var mp *map[string]interface{}
	if mp, err = parseM(&m, 0); err != nil {
		return
	}
	d = *mp
	return
}

// parseM PARSE
func parseM(s *map[string]interface{}, deep int) (d *map[string]interface{}, err error) {
	if deep < ParseMapMax {
		d = s
	} else {
		err = ErrSongoMapDeepOutOf
		return
	}

	// toLower
	for k, v := range *d {
		if _k := strings.ToLower(k); _k != k {
			(*d)[_k] = v
			delete(*d, k)
		}
	}

	// 排序及统计
	var (
		idx   = 0
		lis   = make([]string, len(*s))
		orLen = 0
	)
	for k := range *s {
		lis[idx] = k
		idx++
		if strings.Index(k, TagQueryKeyOr) == 0 {
			orLen++
		}
	}
	sort.Strings(lis)

	for _, k := range lis {
		v := (*s)[k]
		var (
			oper string // 操作符
			comp string // 比较符
			fix  = k    // key
			vFix = v    // val
		)
		oper, comp, fix = keyParse(k)

		switch oper {
		case TagQueryKeyOr, TagQueryKeyAnd:
			// 默认都是数组
			_lis, ok := v.([]interface{})
			if !ok {
				_lis = []interface{}{v}
			}
			// 值转换
			lisFix := []interface{}{}

			for _, _v := range _lis {
				tag := comp
				_vFix := _v

				// 值解析
				switch val := _v.(type) {
				case float64, float32, int, int8, int16, int32, int64:
					if len(fix) > 0 {
						_vFix = map[string]interface{}{fix: val}
					}
				case string:
					// 字符串中可能含有比较符号
					_tag, _valStr := isTagVal(val)
					// 完善结构
					if len(comp) == 0 && len(_tag) == 0 && len(fix) > 0 {
						_vFix = map[string]interface{}{fix: val}
						break
					}
					_vFix = _valStr
					if len(comp) > 0 && len(_tag) > 0 {
						// key与val中均读到比较符
						err = ErrSongoFormatInvalid // 无法识别
						return
					} else if len(_tag) > 0 {
						tag = _tag
					}
				case map[string]interface{}:
					if len(tag) > 0 {
						err = ErrSongoFormatInvalid // 格式有误
						return
					}
					// 如果是map再继续解析
					if mRec, _err := parseM(&val, deep+1); _err == nil {
						_vFix = *mRec
					} else {
						err = _err
						return
					}
				default:
					break
				}

				// 值转换
				if len(tag) > 0 {
					// tag转换
					tag = songoTagMongo(tag)
					// tag: 特殊处理like
					if tag == "$regex" {
						if s, ok := _vFix.(string); ok {
							_vFix = bson.RegEx{s, "i"}
						}
					}
					// key中已经含有比较符
					_v = map[string]interface{}{
						fix: map[string]interface{}{
							tag: _vFix,
						},
					}
				} else {
					_v = _vFix
				}

				// 新值入数组
				if _v != nil {
					lisFix = append(lisFix, _v)
				}
			}

			// 数组替换
			if oper == TagQueryKeyOr && orLen > 1 {
				// 为or加一层and
				v = []interface{}{
					map[string]interface{}{
						songoTagMongo(TagQueryKeyOr): lisFix,
					},
				}
				oper = TagQueryKeyAnd
			} else {
				v = lisFix
			}
		default:
			// 解析表达
			tag := comp

			// 值解析
			switch val := v.(type) {
			case string:
				// 字符串中可能含有比较符号
				_tag, _valStr := isTagVal(val)
				vFix = _valStr
				if len(comp) > 0 && len(_tag) > 0 {
					err = ErrSongoFormatInvalid // 无法识别
					return
				} else if len(_tag) > 0 {
					tag = _tag
				}
				// break
			case map[string]interface{}:
				if len(tag) > 0 {
					err = ErrSongoFormatInvalid // 格式有误
					return
				}
				// 如果是map再继续解析
				if mRec, _err := parseM(&val, deep+1); _err == nil {
					vFix = *mRec
				} else {
					err = _err
					return
				}
				// break
			default:
				break
			}

			// 值转换
			if len(tag) > 0 {
				// tag转换
				tag = songoTagMongo(tag)
				// tag: 特殊处理like
				if tag == "$regex" {
					if s, ok := vFix.(string); ok {
						vFix = bson.RegEx{s, "i"}
					}
				}
				// key中已经含有比较符
				v = map[string]interface{}{
					tag: vFix,
				}
			} else {
				v = vFix
			}

			// break
		}

		// 转义key
		if len(oper) > 0 {
			oper = songoTagMongo(oper)
			// 识别 {"$or$a":["a", "b"], "$or$b":["1", "2"]}
			done := false
			if vExist, ok := (*s)[oper]; ok {
				if vExistSlice, ok := vExist.([]interface{}); ok {
					if vLis, ok := v.([]interface{}); ok {
						vExistSlice = append(vExistSlice, vLis...)
					}
					(*s)[oper] = vExistSlice
					done = true
				}
			}
			if !done {
				(*s)[oper] = v
			}
			if k != oper {
				delete((*s), k)
			}
		} else {
			if fix = songoTagMongo(fix); len(fix) > 0 {
				(*s)[fix] = v
			}
			if k != fix {
				delete((*s), k)
			}
		}
	}
	return
}
