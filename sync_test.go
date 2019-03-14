package orm

import "testing"

type student struct {
	Name string
	Age  int
	ID   string
	Sex  string
}

func Test_Sync(t *testing.T) {
	if _d, _err := StructModelInfo(student{}); _err == nil {
		if _d == nil {
			t.Fatal("writer and read is diff")
		}
	} else {
		t.Fatal(_err)
	}
	if _d, _err := StructModelInfoNoPrimary(student{}); _err == nil {
		if _d == nil {
			t.Fatal("writer and read is diff")
		}
	} else {
		t.Fatal(_err)
	}
}
