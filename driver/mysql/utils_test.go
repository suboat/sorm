package mysql

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type TestTimeFill struct {
	TimeV1 time.Time
	TimeP1 *time.Time
}

type testTimeFill2 struct {
	*TestTimeFill
	TimeV2 time.Time
	TimeP2 *time.Time
}

type testTimeFill3 struct {
	TestTimeFill
	TimeV3 time.Time
}

//
func Test_PubTimeFill(t *testing.T) {
	as := require.New(t)
	as.Nil(nil)
	var (
		cmp = &time.Time{}
		t1  = new(TestTimeFill)
		t2  = new(testTimeFill2)
		t3  = new(testTimeFill3)
	)
	t2.TestTimeFill = new(TestTimeFill)
	t2.TimeP1 = &time.Time{}
	t2.TimeP2 = &time.Time{}
	t3.TimeP1 = &time.Time{}
	//
	as.Nil(PubTimeFill(t1))
	as.True(t1.TimeV1.Unix() > cmp.Unix())
	t.Logf(`t1.v1 %v`, t1.TimeV1)
	//
	as.Nil(PubTimeFill(t2))
	as.True(t2.TimeV2.Unix() > cmp.Unix())
	as.True(t2.TimeP1.Unix() > cmp.Unix())
	as.True(t2.TimeP2.Unix() > cmp.Unix())
	t.Logf(`t2.v2 %v`, t2.TimeV2)
	t.Logf(`t2.p1 %v`, t2.TimeP1)
	t.Logf(`t2.p2 %v`, t2.TimeP2)
	//
	as.Nil(PubTimeFill(t3))
	as.True(t3.TimeV1.Unix() > cmp.Unix())
	as.True(t3.TimeP1.Unix() > cmp.Unix())
	as.True(t3.TimeV3.Unix() > cmp.Unix())
	t.Logf(`t3.v1 %v`, t3.TimeV1)
	t.Logf(`t3.p1 %v`, t3.TimeP1)
	t.Logf(`t3.v3 %v`, t3.TimeV3)
}
