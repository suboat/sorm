package orm

import (
	"github.com/shopspring/decimal"

	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	// RegValTypeInt 判断字符串是否为Int
	RegValTypeInt = regexp.MustCompile(`^[+-]?[0-9]+$`)
	// RegValTypeNum 判断字符串是否为数字
	RegValTypeNum = regexp.MustCompile(`^[+-]?([0-9]*[.])?[0-9]+$`)
	// RegValTypeTime 判断字符串是否为时间 2016-08-22 11:30:47.269195843 +0800 CST
	RegValTypeTime = regexp.MustCompile(
		`^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:([0-9]*[.])?[0-9]+ [+-][0-9]{4}\s*([A-Z]+)?$`)
	// NormalBeginData 时间类型处理 一个自然的起始时间
	NormalBeginData = time.Date(1950, 1, 1, 0, 0, 0, 0, time.Now().Location())
	// ZoneOffset 时差标记(服务器)
	ZoneOffset = ""
	// CfgFloatAutoRound 浮点数默认保留位数
	CfgFloatAutoRound int32 = 8
)

// PubTimeQueryFormat 日期作为搜索参数时候需要做的转化
func PubTimeQueryFormat(t time.Time) (s string) {
	s = t.String()
	s = s[0 : len(s)-4]
	return
}

// PubTimeStrParse 将时间字符串转为时间格式
func PubTimeStrParse(s string) (t time.Time, err error) {
	if RegValTypeTime.MatchString(s) {
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

// FloatAdd 浮点数加
func FloatAdd(a float64, vals ...float64) (c float64) {
	n := decimal.NewFromFloat(a)
	for _, v := range vals {
		n = n.Add(decimal.NewFromFloat(v))
	}
	c, _ = n.Float64()
	return
}

// FloatAddRound 浮点数加,并保留有效位数
func FloatAddRound(a float64, vals ...float64) (c float64) {
	c = FloatRoundAuto(FloatAdd(a, vals...))
	return
}

// FloatSub 浮点数减
func FloatSub(a float64, vals ...float64) (c float64) {
	n := decimal.NewFromFloat(a)
	for _, v := range vals {
		n = n.Sub(decimal.NewFromFloat(v))
	}
	c, _ = n.Float64()
	return
}

// FloatSubRound 浮点数减,并保留有效位数
func FloatSubRound(a float64, vals ...float64) (c float64) {
	c = FloatRoundAuto(FloatSub(a, vals...))
	return
}

// FloatMul 浮点数乘
func FloatMul(a float64, vals ...float64) (c float64) {
	n := decimal.NewFromFloat(a)
	for _, v := range vals {
		n = n.Mul(decimal.NewFromFloat(v))
	}
	c, _ = n.Float64()
	return
}

// FloatMulRound 浮点数乘,并保留有效位数
func FloatMulRound(a float64, vals ...float64) (c float64) {
	c = FloatRoundAuto(FloatMul(a, vals...))
	return
}

// FloatDiv 浮点数除
func FloatDiv(a float64, vals ...float64) (c float64) {
	n := decimal.NewFromFloat(a)
	for _, v := range vals {
		n = n.Div(decimal.NewFromFloat(v))
	}
	c, _ = n.Float64()
	return
}

// FloatDivRound 浮点数除,并保留有效位数
func FloatDivRound(a float64, vals ...float64) (c float64) {
	c = FloatRoundAuto(FloatDiv(a, vals...))
	return
}

// FloatRound 浮点数保留位数
func FloatRound(a float64, r int32) (c float64) {
	n := decimal.NewFromFloat(a).Round(r)
	c, _ = n.Float64()
	return
}

// FloatRoundAuto 浮点数保留系统默认位数
func FloatRoundAuto(a float64) (c float64) {
	return FloatRound(a, CfgFloatAutoRound)
}

// JSONMust 转为json
func JSONMust(inf interface{}) (r string) {
	r = "{}"
	if inf == nil {
		return
	}
	switch v := inf.(type) {
	case string:
		if len(v) > 2 && v[0] == '{' && v[len(v)-1] == '}' {
			r = v
		}
	case *string:
		_v := *v
		if len(_v) > 2 && _v[0] == '{' && _v[len(_v)-1] == '}' {
			r = _v
		}
	default:
		if r1, _err := json.Marshal(inf); _err == nil {
			r = string(r1)
		}
	}
	return
}

// 解析并创建一个实例
func ReflectElemNew(a interface{}) (ret interface{}) {
	if t := reflectElemType(a); t != nil {
		ret = reflect.New(t).Interface()
	}
	return
}
func reflectElemType(a interface{}) reflect.Type {
	for t := reflect.TypeOf(a); ; {
		switch t.Kind() {
		case reflect.Ptr, reflect.Slice:
			t = t.Elem()
		default:
			return t
		}
	}
}

// init
func initUtils() (err error) {
	// ZoneOffset on server
	if _, offset := time.Now().Zone(); offset > 0 {
		ZoneOffset = fmt.Sprintf("+%02d:00", offset/60/60)
	} else {
		ZoneOffset = fmt.Sprintf("-%02d:00", offset/60/60)
	}
	return
}
