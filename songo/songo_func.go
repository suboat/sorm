package songo

import (
	"strings"
)

// valid
func isSongoMapValid(m map[string]interface{}) (err error) {
	if m == nil {
		err = ErrSongoMapUndefined
		return
	}
	return
}

// 判断一个字符串是否含tagVal, 有则返回
func isTagVal(s string) (tag string, val string) {
	if val = s; len(val) <= 2 || val[0] != TagSep {
		return
	}
	if idx := strings.Index(s[1:], string(TagSep)); idx > -1 {
		tag = s[0 : idx+2]
		val = s[idx+2:]
	}
	return
}

// 判断一个key中是否有操作符
func isKeyOper(s string) (oper string, key string) {
	if key = s; len(key) <= 2 || key[0] != TagSep {
		return
	}
	if idx := strings.Index(s[1:], string(TagSep)); idx > -1 {
		oper = s[0 : idx+2]
		key = s[idx+2:]
	}
	return
}

// 判断一个key中是否有比较符号
func isKeyComp(s string) (comp string, key string) {
	if key = s; len(key) <= 2 || key[len(key)-1] != TagSep {
		return
	}
	if idx := strings.Index(s[1:], string(TagSep)); idx > -1 {
		hIdx := -1
		for i := len(key) - 2; i > -1; i-- {
			if key[i] == TagSep {
				hIdx = i
				break
			}
		}
		if hIdx > -1 {
			if hIdx == 0 {
				key = ""
				return
			}
			comp = key[hIdx:]
			key = key[0:hIdx]
		}
	}
	return
}

// 解析一个key中的操作符，比较符
func keyParse(s string) (oper string, comp string, key string) {
	oper, key = isKeyOper(s)
	comp, key = isKeyComp(key)
	return
}

// mapDeepCopy 复制map
//func mapDeepCopy(value interface{}) interface{} {
//	if valueMap, ok := value.(map[string]interface{}); ok {
//		newMap := make(map[string]interface{})
//		for k, v := range valueMap {
//			newMap[k] = mapDeepCopy(v)
//		}
//		return newMap
//	} else if valueSlice, ok := value.([]interface{}); ok {
//		newSlice := make([]interface{}, len(valueSlice))
//		for k, v := range valueSlice {
//			newSlice[k] = mapDeepCopy(v)
//		}
//		return newSlice
//	}
//	return value
//}
