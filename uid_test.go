package orm

import (
	"testing"
)

func Test_Uid(t *testing.T) {
	s := NewUid()
	t.Log("uid ", s, "len:", len(s))
}
