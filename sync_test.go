package orm

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type student struct {
	Name string
	Age  int
	ID   string
	Sex  string
}

func Test_Sync(t *testing.T) {
	as := require.New(t)
	if true {
		_d, _err := StructModelInfo(student{})
		as.Nil(_err, "writer and read is diff")
		as.Equal(4, len(_d))
		for _, v := range _d {
			t.Logf(`[test-sync-field] %s`, v.Name)
		}
	}
	if true {
		_d, _err := StructModelInfoNoPrimary(student{})
		as.Nil(_err, "writer and read is diff")
		as.Equal(4, len(_d))
		for _, v := range _d {
			t.Logf(`[test-sync-field] %s`, v.Name)
		}
	}
	if true {
		var dl []*student
		_d, _err := StructModelInfoByDest(&dl)
		as.Nil(_err, "writer and read is diff")
		as.Equal(4, len(_d))
		for _, v := range _d {
			t.Logf(`[test-sync-field] %s`, v.Name)
		}
	}
}
