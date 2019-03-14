package songo

import (
	"strings"
)

// SortSafe 按照白名单，取排序值
func SortSafe(whiteList, defaults []string, inputs interface{}) (result []string) {
	if whiteList == nil {
		// defaults || inputs
		if defaults != nil {
			result = defaults
		}
		if inputs != nil {
			if _inputs, ok := inputs.([]string); ok {
				result = _inputs
			}
		}
		return
	}

	var (
		validMap = make(map[string]bool)
	)
	for _, _v := range whiteList {
		v := strings.ToLower(_v)
		if strings.HasPrefix(_v, "+") || strings.HasPrefix(_v, "-") {
			v = v[1:]
		}
		validMap[v] = true
		validMap["+"+v] = true
		validMap["-"+v] = true
	}
	if inputs != nil {
		if _inputs, ok := inputs.([]string); ok {
			for _, v := range _inputs {
				if _, ok := validMap[strings.ToLower(v)]; ok {
					result = append(result, v)
				}
			}
		}
	}

	if result == nil {
		result = defaults
	}
	return
}

// ParseSafe 按照白名单，黑名单，默认值将map过滤
func ParseSafe(m, whiteList, blackList, defaultVals map[string]interface{}) (err error) {
	// valid
	if m == nil {
		return
	} else if whiteList == nil && blackList == nil && defaultVals == nil {
		return
	}

	// 白名单，只保留这些
	if whiteList != nil {
		if err = parseSafeFilter(false, m, whiteList, 0); err != nil {
			return
		}
	}

	// 黑名单，从中去除
	if blackList != nil {
		if err = parseSafeFilter(true, m, blackList, 0); err != nil {
			return
		}
	}

	// 默认值，设置一级数值默认值
	for k, v := range defaultVals {
		m[k] = v
	}

	return
}

func parseSafeFilter(isRemove bool, m map[string]interface{}, filterMap map[string]interface{}, deep int) (err error) {

	for k, v := range m {
		var (
			oper string // 操作符
			//comp string // 比较符
			fix = k // key
			//vFix = v    // val
		)
		oper, _, fix = keyParse(k)

		if isRemove {
			// 键分析: 剔除
			if _, ok := filterMap[fix]; ok {
				delete(m, k)
				continue
			}
		}

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
				// 值解析
				switch val := _v.(type) {
				case string:
					break
				case map[string]interface{}:
					// 如果是map再继续解析
					if err = parseSafeFilter(isRemove, val, filterMap, deep+1); err != nil {
						return
					} else if len(val) > 0 {
						_v = val
					} else {
						_v = nil
					}
					// break
				default:
					break
				}

				// 新值入数组
				if _v != nil {
					lisFix = append(lisFix, _v)
				}
			}

			// 数组替换
			v = lisFix
			if len(lisFix) > 0 {
				m[k] = lisFix
			} else {
				delete(m, k)
			}
			// break
		default:
			// 键分析: 去除一级结构中的无意义值
			if (deep == 0) && (len(fix)) == 0 {
				delete(m, k)
				break
			}

			if !isRemove {
				// 白名单模式
				if _, ok := filterMap[fix]; !ok {
					delete(m, k)
					break
				}
			}

			// 值解析
			switch val := v.(type) {
			case string:
				break
			case map[string]interface{}:
				if err = parseSafeFilter(isRemove, val, filterMap, deep+1); err != nil {
					return
				}
				// break
			default:
				break
			}
			// break
		}
	}

	return
}
