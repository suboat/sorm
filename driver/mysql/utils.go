package mysql

import (
	"github.com/suboat/sorm"

	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	// 时间格式转换
	RegTimeWithZone = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}.*`)
)

//
func PubFieldWrap(s string) (ret string) {
	return fmt.Sprintf(`"%s"`, strings.ToLower(s))
}

//
func PubFieldWrapAll(s []string) (ret []string) {
	for _, v := range s {
		ret = append(ret, PubFieldWrap(v))
	}
	return
}

//
func PubFieldWrapByFieldInfo(s []*orm.FieldInfo) (ret []string) {
	for _, v := range s {
		ret = append(ret, PubFieldWrap(v.Name))
	}
	return
}

//
func PubFieldWrapByDest(dest interface{}) (ret []string) {
	if _ret, _err := orm.StructModelInfoByDest(dest); _err == nil {
		return PubFieldWrapByFieldInfo(_ret)
	}
	return
}

// PubTimeConvert 转换为UTC时间
func PubTimeConvert(in string) string {
	if RegTimeWithZone.MatchString(in) == false {
		return in
	}
	if _t, _err := time.Parse(time.RFC3339Nano, in); _err == nil {
		return _t.UTC().Format("2006-01-02 15:04:05.999999999")
	}
	return in
}

// 遍历结构的0时间,赋值
func PubTimeFill(st interface{}) (err error) {
	if st == nil {
		return
	}
	var (
		stVal        = reflect.Indirect(reflect.ValueOf(st))
		stKind       = stVal.Kind()
		stType       = stVal.Type()
		offset int64 = -60 * 60 * 24
	)
	if stKind != reflect.Struct {
		return
	}
	// 寻找时间值
	for i := 0; i < stType.NumField(); i++ {
		var (
			fType = stType.Field(i)
			fVal  = stVal.Field(i)
			c     = int(rune(fType.Name[0]))
		)
		if !(c >= 'A' && c <= 'Z') && (!fType.Anonymous) {
			continue
		}
		switch fType.Type.Kind() {
		case reflect.Struct:
			if fType.Anonymous {
				if err = PubTimeFill(fVal.Addr().Interface()); err != nil {
					return
				}
			} else {
				// time type and others
				switch v := fVal.Interface().(type) {
				case time.Time:
					// 更新0值
					if v.Unix() < offset {
						if fVal.CanAddr() {
							var _v = fVal.Addr().Interface().(*time.Time)
							*_v = time.Unix(offset, 0)
						}
					}
				}
			}
		case reflect.Ptr:
			if fType.Anonymous {
				// inherit
				if fVal.IsNil() {
					continue
				}
				//
				if err = PubTimeFill(fVal.Interface()); err != nil {
					return
				}
			} else {
				switch v := fVal.Interface().(type) {
				case *time.Time:
					// 更新0值
					if v != nil && v.Unix() < offset {
						*v = time.Unix(offset, 0)
					}
				}
			}
		default:
			break
		}
	}
	return
}
