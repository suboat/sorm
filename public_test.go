package orm

import (
	"testing"
	"time"
)

// 测试日期格式转换
func Test_TimeConv(t *testing.T) {
	var (
		now    = time.Now()
		nowStr = PubTimeQueryFormatP(&now)
	)
	t.Logf("PubTimeQueryFormat: %s (%s)", nowStr, now.String())
	if tFmt, err := PubTimeStrParse(nowStr); err != nil {
		t.Fatal(err)
		return
	} else {
		t.Logf("PubTimeStrParse: %s (%s)", tFmt.String(), nowStr)
	}
	return
}
