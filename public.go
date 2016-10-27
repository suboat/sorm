package orm

import (
	"fmt"
	"strings"
	"time"
)

var (
	NormalBeginData = time.Date(1950, 1, 1, 0, 0, 0, 0, time.Now().Location()) // 一个自然的起始时间
	ZoneOffset      = ""                                                       // 时差标记(服务器)
)

// 日期作为搜索参数时候需要做的转化
func PubTimeQueryFormat(t time.Time) (s string) {
	return PubTimeQueryFormatP(&t)
}

// point
func PubTimeQueryFormatP(t *time.Time) (s string) {
	s = t.String()
	s = s[0 : len(s)-4]
	return
}

// 将时间字符串转为时间格式
func PubTimeStrParse(s string) (t time.Time, err error) {
	//
	if RegValTypeTime.MatchString(s) == true {
		var (
			lis = strings.Split(s, " ")
			//yearStr   = s[0:4]
			//monthStr  = s[4:6]
			//dayStr    = s[6:8]
			//hourStr   = s[8:10]
			//minuteStr = s[10:12]
			//secondStr = s[12:14]
			//s         = fmt.Sprintf("%s-%s-%sT%s:%s:%s%s", yearStr, monthStr, dayStr, hourStr, minuteStr, secondStr,
			//	ZoneOffset)
			tStr = fmt.Sprintf("%sT%s%s:00", lis[0], lis[1], lis[2][0:3])
		)
		t, err = time.Parse(time.RFC3339, tStr)
		return
	}

	return
}
