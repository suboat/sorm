package orm

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	RegFieldTag = regexp.MustCompile(`(\w+)(\((\w+)\))?`)
)

// 结构体中字段信息
type FieldInfo struct {
	TableName  string
	Name       string
	Kind       string
	Primary    bool
	Index      bool
	Serial     bool
	IndexKeys  []string
	Unique     bool
	UniqueKeys []string
	Size       int64
}

// 将结构体中的字段转为map映射，供搜索用。目前只支持两层潜逃内的string
// TODO: 优化
func filedToMap(st interface{}, m *map[string]interface{}, d int) (err error) {
	stVal := reflect.Indirect(reflect.ValueOf(st))
	d += 1
	switch stVal.Kind() {
	case reflect.Struct:
		stType := stVal.Type()
		for i := 0; i < stType.NumField(); i++ {
			fType := stType.Field(i)
			fVal := stVal.Field(i)
			c := int(rune(fType.Name[0]))

			//println("debug", fType.Name, fVal.Kind().String(), fType.Tag.Get(orm.TagKey))
			if c >= 'A' && c <= 'Z' {
				key := strings.ToLower(fType.Name)
				if _, ok := (*m)[key]; ok == false {
					if fVal.Kind() == reflect.Struct {
						if d < 2 {
							filedToMap(fVal.Interface(), m, d)
						}
					} else {
						switch fVal.Kind() {
						case reflect.String:
							val := fVal.String()
							if len(val) > 0 {
								(*m)[key] = fVal.String()
							}
							break
						default:
							break
						}
					}
				}
			}

		}
		break
	default:
		break
	}
	return
}

func StructMap(st interface{}) (M, error) {
	var (
		m   = make(map[string]interface{})
		err error
	)

	err = filedToMap(st, &m, 0)
	return M(m), err
}

// 结构体提取model信息
func StructModelInfo(st interface{}) (res []*FieldInfo, err error) {
	res = []*FieldInfo{}
	namMap := make(map[string]int) // embed overwtire
	return structModelInfo(st, &res, &namMap)
}
func structModelInfo(st interface{}, src *[]*FieldInfo, m *map[string]int) (res []*FieldInfo, err error) {
	if src != nil {
		res = *src
	} else {
		res = []*FieldInfo{}
	}
	stVal := reflect.Indirect(reflect.ValueOf(st))
	switch stVal.Kind() {
	case reflect.Struct:
		stType := stVal.Type()
		tbName := strings.ToLower(stType.Name())
		for i := 0; i < stType.NumField(); i++ {
			var (
				fType = stType.Field(i)
				fVal  = stVal.Field(i)
				info  = new(FieldInfo)
				fKey  = fType.Tag.Get(TagKey)
				c     = int(rune(fType.Name[0]))
			)
			// ignore tag
			if fKey == "-" {
				continue
			}
			//println("aaa", fType.Anonymous, fType.Name)

			// can not export and not inherit
			if !(c >= 'A' && c <= 'Z') && (fType.Anonymous == false) {
				continue
			}
			// parser name of column
			info.TableName = tbName
			info.Name = strings.ToLower(fType.Name)
			info.IndexKeys = []string{info.Name}
			info.UniqueKeys = []string{info.Name}
			//println(info.Name, fType.Type.Kind().String())
			// parser type of column
			switch fType.Type.Kind() {
			case reflect.String:
				info.Kind = "text"
				break
			case reflect.Int:
				info.Kind = "integer"
				break
			case reflect.Int64:
				info.Kind = "bigint"
				break
			case reflect.Uint:
				info.Kind = "unit"
				break
			case reflect.Float32:
				info.Kind = "float"
				break
			case reflect.Float64:
				info.Kind = "float"
				break
			case reflect.Bool:
				info.Kind = "boolean"
				break
			case reflect.Struct:
				// inherit
				if fType.Anonymous == true {
					if res, err = structModelInfo(fVal.Interface(), &res, m); err != nil {
						return
					}
				} else {
					// time type
					if _, ok := fVal.Interface().(time.Time); ok {
						//info.Kind = "date"
						info.Kind = "timestamp"
					} else {
						// jsonb
						info.Kind = "json"
					}
				}
				break
			case reflect.Slice:
				//info.Kind = "json"
				if _, ok := fVal.Interface().([]byte); ok {
					info.Kind = "bytearray"
				} else {
					info.Kind = "json"
				}
				break
			default:
				continue
				break
			}
			if len(info.Kind) == 0 {
				continue
			}
			// parser tag
			for _, vStr := range strings.Split(fKey, ";") {
				for _, s := range strings.Split(vStr, " ") {
					r := RegFieldTag.FindStringSubmatch(s)
					if len(r) == 0 {
						continue
					}
					k := r[1]
					v := r[3]
					// parser
					//println("debug", tbName, info.Name, k, v)
					if k == "primary" {
						info.Primary = true
					} else if k == "index" {
						info.Index = true
						for _, _key := range strings.Split(v, ",") {
							if len(_key) > 0 {
								info.IndexKeys = append(info.IndexKeys, _key)
							}
						}
					} else if k == "unique" {
						info.Unique = true
						for _, _key := range strings.Split(v, ",") {
							if len(_key) > 0 {
								info.UniqueKeys = append(info.UniqueKeys, _key)
							}
						}
					} else if k == "size" {
						i := 0
						if i, err = strconv.Atoi(v); err != nil {
							return
						}
						info.Size = int64(i)
						if info.Kind == "text" {
							info.Kind = "varchar"
						}
					} else if k == "serial" {
						if info.Kind == "bigint" {
							info.Kind = "bigserial"
							info.Serial = true
						} else if info.Kind == "integer" {
							info.Kind = "serial"
							info.Serial = true
						}
						if info.Serial == true {
							info.Primary = true
						}
					}
				}
			}
			// final
			if idx, ok := (*m)[info.Name]; ok == true {
				// overwrite
				res[idx] = info
			} else {
				(*m)[info.Name] = len(res)
				res = append(res, info)
			}
			//println(i, fType.Name, fType.Type.Kind().String(), info.Index)
		}
		break
	default:
		break
	}
	return
}

// 判断一个字符串是否含tagVal, 有则返回
func IsTagVal(s string) (tag string, val string) {
	if len(s) <= 2 || s[0] != '%' {
		return
	}

	if idx := strings.Index(s[1:], `%`); idx > -1 {
		tag = s[0 : idx+2]
		val = s[idx+2:]
	}

	return
}
