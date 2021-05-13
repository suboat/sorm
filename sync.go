package orm

import (
	"github.com/suboat/sorm/types"

	"database/sql/driver"
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 将结构体同步至数据库
var (
	// RegFieldTag 匹配规则
	RegFieldTag = regexp.MustCompile(`(\w+)(\(([\w,]+)\))?`)
)

// FieldInfo 结构体中字段信息
type FieldInfo struct {
	ReflectKind uint        //
	TableName   string      //
	Name        string      //
	Kind        string      //
	Primary     bool        //
	PrimaryKeys []string    //
	Index       bool        //
	Serial      bool        //
	IndexKeys   []string    //
	Unique      bool        //
	UniqueKeys  []string    //
	Size        int         //
	Precision   int         //
	IndexText   bool        //
	DefaultVal  interface{} //
	AllowNull   bool        //
}

// 将结构体中的字段转为map映射，供搜索用。目前只支持两层嵌套内的string TODO: 优化
//func filedToMap(st interface{}, m *map[string]interface{}, d int) (err error) {
//	stVal := reflect.Indirect(reflect.ValueOf(st))
//	d++
//	switch stVal.Kind() {
//	case reflect.Struct:
//		stType := stVal.Type()
//		for i := 0; i < stType.NumField(); i++ {
//			fType := stType.Field(i)
//			fVal := stVal.Field(i)
//			c := int(rune(fType.Name[0]))
//
//			//println("debug", fType.Name, fVal.Kind().String(), fType.Tag.Get(orm.TagKey))
//			if c >= 'A' && c <= 'Z' {
//				key := strings.ToLower(fType.Name)
//				if _, ok := (*m)[key]; ok == false {
//					if fVal.Kind() == reflect.Struct {
//						if d < 2 {
//							filedToMap(fVal.Interface(), m, d)
//						}
//					} else {
//						switch fVal.Kind() {
//						case reflect.String:
//							val := fVal.String()
//							if len(val) > 0 {
//								(*m)[key] = fVal.String()
//							}
//							break
//						default:
//							break
//						}
//					}
//				}
//			}
//
//		}
//		break
//	default:
//		break
//	}
//	return
//}

// StructModelInfo 结构体提取model信息
func StructModelInfo(st interface{}) (res []*FieldInfo, err error) {
	res = []*FieldInfo{}
	return structModelInfo(st, &res, nil)
}

// StructModelInfoNoPrimary 结构体提取model信息: 将主键描述解读为unique, 适用于不支持主键定义的数据库驱动使用
func StructModelInfoNoPrimary(st interface{}) (res []*FieldInfo, err error) {
	res = []*FieldInfo{}
	primary := ""
	return structModelInfo(st, &res, &primary)
}

//
func structModelInfo(st interface{}, src *[]*FieldInfo, primary *string) (res []*FieldInfo, err error) {
	if src != nil {
		res = *src
	} else {
		res = []*FieldInfo{}
	}
	var (
		primaryKeys []string
		primaryMap  = make(map[string]int)
	)
	//
	stVal := reflect.Indirect(reflect.ValueOf(st))
	switch stVal.Kind() {
	case reflect.Struct: // 结构体
		stType := stVal.Type()                   // 结构体种类
		tbName := strings.ToLower(stType.Name()) // 结构体名称(大部分为表名)
		for i := 0; i < stType.NumField(); i++ { // 根据结构体的变量数量遍历
			var (
				fType = stType.Field(i) // 变量类型 如: {UID  string sorm:"size(36);unique" json:"uid" 8 [1] false}
				fVal  = stVal.Field(i)  // 字段的默认值
				info  = new(FieldInfo)
				fKey  = strings.ToLower(fType.Tag.Get(OrmKey))     // key 如:size(36);unique
				dKey  = strings.Split(fType.Tag.Get("db"), ",")[0] // alias field name 使用db字眼的时候才有
				c     = int(rune(fType.Name[0]))                   // 字段的首字母的ascii值
			)
			info.AllowNull = true // 默认允许空

			// ignore tag 为"-"的不做处理(不当作表的字段)
			if fKey == "-" || dKey == "-" {
				continue
			}
			// can not export and not inherit
			if !(c >= 'A' && c <= 'Z') && (!fType.Anonymous) {
				continue
			}
			//Log.Debugf(`[struct-info] %v %v`, fType.Anonymous, fType.Name)

			// 赋值
			// parser name of column
			info.TableName = tbName
			info.Name = strings.ToLower(fType.Name)
			info.PrimaryKeys = []string{info.Name}
			info.IndexKeys = []string{info.Name}
			info.UniqueKeys = []string{info.Name}
			// 如果dkey存在则以dkey为字段名
			if len(dKey) > 0 {
				info.Name = dKey
			}
			//Log.Debugf(`[struct-info] %s %s`, info.Name, fType.Type.Kind().String())

			// parser type of column  对字段的数据类型进行转换
			info.ReflectKind = uint(fType.Type.Kind())
			switch fType.Type.Kind() {
			case reflect.String:
				info.Kind = "text"
				info.DefaultVal = ""
			case reflect.Int, reflect.Int8, reflect.Int32:
				info.Kind = "integer"
				info.DefaultVal = 0
			case reflect.Uint, reflect.Uint8, reflect.Uint32:
				info.Kind = "unit"
				info.DefaultVal = 0
			case reflect.Int64, reflect.Uint64:
				info.Kind = "bigint"
				info.DefaultVal = 0
			case reflect.Float32:
				info.Kind = "float"
				info.DefaultVal = 0
			case reflect.Float64:
				info.Kind = "float"
				info.DefaultVal = 0
			case reflect.Bool:
				info.Kind = "boolean"
				info.DefaultVal = false
			case reflect.Slice:
				//info.Kind = "json"
				switch fVal.Interface().(type) {
				case []byte,
					types.SliceInf, types.SliceStr, // slice type
					json.RawMessage, // json type
					types.JSONText:  // package type
					info.Kind = "bytearray"
					info.DefaultVal = ""
				default:
					info.Kind = "json"
					info.DefaultVal = "[]"
				}
			case reflect.Map:
				// jsonb
				info.Kind = "json"
				info.DefaultVal = "{}"
			case reflect.Struct:
				// inherit
				if fType.Anonymous { // 是继承的结构体(表与表之间,无key的那种)
					if res, err = structModelInfo(fVal.Interface(), &res, primary); err != nil {
						return
					}
				} else {
					// time type and others
					switch fVal.Interface().(type) {
					case time.Time:
						//info.Kind = "date"
						info.Kind = "timestamp"
						info.DefaultVal = DefaultTimeStr // default time
					case types.BigInt:
						info.Kind = "bigint"
						info.DefaultVal = 0
					case driver.Valuer:
						// jsonb
						info.Kind = "json"
						info.DefaultVal = "{}"
					default:
						//
					}
				}
			case reflect.Ptr:
				if fType.Anonymous {
					// inherit
					if fVal.IsNil() {
						// TODO:
						err = ErrSyncEmbedPointNil
						return
					}
					if res, err = structModelInfo(fVal.Interface(), &res, primary); err != nil {
						return
					}
				} else {
					// null:常见类型
					info.DefaultVal = nil
					switch fVal.Interface().(type) {
					case *string:
						// 可控字段
						info.Kind = "text"
					case *int, *int16, *int32, *int64:
						info.Kind = "integer"
					case *float64, *float32:
						info.Kind = "float"
					case *bool:
						info.Kind = "boolean"
					case *time.Time:
						info.Kind = "timestamp"
					default:
						// 指针类型的field不处理
						// json
						//info.Kind = "json"
						//info.DefaultVal = "{}"
						continue
					}
				}
			default:
				// ignore
				continue
			}
			if len(info.Kind) == 0 {
				continue
			}

			// parser tag
			for _, vStr := range strings.Split(fKey, ";") { // key里面的处理 如: "index;size(36)"
				for _, s := range strings.Split(vStr, " ") {
					r := RegFieldTag.FindStringSubmatch(s)
					if len(r) == 0 {
						continue
					}
					k := r[1]
					v := r[3]

					// fallback primary to unique
					if (k == "primary") && (primary != nil) {
						k = "unique"
					}

					// parser
					//println("debug", tbName, info.Name, k, v)
					switch k {
					case "primary":
						info.Primary = true
						info.AllowNull = false
						for _, _key := range strings.Split(v, ",") {
							if len(_key) > 0 {
								info.PrimaryKeys = append(info.PrimaryKeys, _key)
								primaryMap[_key] += 1
							}
						}
						_primary := strings.Join(info.PrimaryKeys, ",")
						primary = &_primary // just not nil
						primaryKeys = info.PrimaryKeys
					case "serial":
						if info.Kind == "bigint" {
							info.Kind = "bigserial"
							info.Serial = true
						} else if info.Kind == "integer" {
							info.Kind = "serial"
							info.Serial = true
						}
					case "unique":
						info.Unique = true
						for _, _key := range strings.Split(v, ",") {
							if len(_key) > 0 {
								info.UniqueKeys = append(info.UniqueKeys, _key)
							}
						}
					case "index":
						info.Index = true
						for _, _key := range strings.Split(v, ",") {
							if len(_key) > 0 {
								info.IndexKeys = append(info.IndexKeys, _key)
							}
						}
					case "size":
						i := 0
						if i, err = strconv.Atoi(v); err != nil {
							return
						}
						info.Size = i
						if info.Kind == "text" {
							info.Kind = "varchar"
						}
					case "char":
						// rewrite type
						if info.Kind == "text" || info.Kind == "varchar" {
							info.Kind = "char"
							info.DefaultVal = ""
							if i, er := strconv.Atoi(v); er == nil {
								info.Size = i
							}
						}
					case "json":
						info.Kind = "json"
						info.DefaultVal = "{}"
					case "decimal", "numeric":
						var args []int
						for _, _key := range strings.Split(v, ",") {
							if _v, _err := strconv.Atoi(_key); _err == nil {
								args = append(args, _v)
							} else {
								err = _err
								return
							}
						}
						if len(args) >= 2 {
							info.Size = args[0]
							info.Precision = args[1]
						} else if len(args) == 1 {
							info.Size = args[0]
							info.Precision = -1 // 未定义精度
						} else {
							info.Precision = -1 // 未定义精度
						}
						info.Kind = "decimal"
					default:
					}
				}
			}

			// final
			res = append(res, info)
			//println(i, fType.Name, fType.Type.Kind().String(), info.Index)
		}
	default:
		break
	}
	// 补充其它字段的主键信息
	if len(primaryKeys) > 0 {
		for _, v := range res {
			if v.Primary == false {
				v.PrimaryKeys = primaryKeys
				if primaryMap[v.Name] > 0 {
					v.AllowNull = false
				}
			}
		}
	}
	return
}
