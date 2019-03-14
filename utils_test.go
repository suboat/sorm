package orm

import (
	"testing"

	"time"
)

//
func Test_Float_Utils(t *testing.T) {
	if n := FloatAdd(10.1, 0.4, 0.5); n != 11 {
		t.Fatalf("get %v", n)
	}
	if n := FloatSub(10.1, 0.1, 1.0); n != 9 {
		t.Fatalf("get %v", n)
	}
	if n := FloatMul(10.0, 10.0, 10.0); n != 1000 {
		t.Fatalf("get %v", n)
	}
	if n := FloatDiv(10.0, 10.0, 10.0); n != 0.1 {
		t.Fatalf("get %v", n)
	}
	if n := FloatAddRound(10.0, 0.2); n != 10.2 {
		t.Fatalf("get %v", n)
	}
	if n := FloatSubRound(10.0, 0.02); n != 9.98 {
		t.Fatalf("get %v", n)
	}
	if n := FloatMulRound(10.0, 0.2); n != 2 {
		t.Fatalf("get %v", n)
	}
	if n := FloatDivRound(10.0, 3); n != 3.33333333 {
		t.Fatalf("get %v", n)
	}
	if n := FloatRound(3.1415954787, 4); n != 3.1416 {
		t.Fatalf("get %v", n)
	}
	if n := FloatRoundAuto(3.1415954787); n != 3.14159548 {
		t.Fatalf("get %v", n)
	}
}

//
func Test_Time(t *testing.T) {
	t.Logf(PubTimeQueryFormat(time.Now()))
	if te, err := PubTimeStrParse("2019-02-21 10:05:21.944301352 +0800 CST m=+0.00065"); err != nil {
		t.Fatal(err)
	} else {
		t.Log(te)
	}
}

//
func TestJSONMust(t *testing.T) {
	a := map[string]interface{}{
		"id": 10,
	}
	t.Log(JSONMust(a))
}
