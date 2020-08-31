package orm

import (
	"testing"
	"time"
)

// 测试日期格式转换
func Test_TimeConv(t *testing.T) {
	var (
		now    = time.Now()
		nowStr = PubTimeQueryFormat(now)
	)
	t.Logf("PubTimeQueryFormat: %s (%s)", nowStr, now.String())
	if tFmt, err := PubTimeStrParse(nowStr); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("PubTimeStrParse: %s (%s)", tFmt.String(), nowStr)
	}
}

//
func TestNew(t *testing.T) {
	t.Log(New("", ""))
}

//
func TestSetLogLevel(t *testing.T) {
	var a = 10
	SetLogLevel(a)
}
func Test_Other(t *testing.T) {
	t.Log(defaultHookParseSafe(nil, nil, nil, nil))
	t.Log(defaultHookParseMgo(map[string]interface{}{"id": 10}))
}
