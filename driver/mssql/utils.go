package mssql

import (
	"regexp"
	"time"
)

var (
	// 时间格式转换
	RegTimeWithZone = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}.*`)
)

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
