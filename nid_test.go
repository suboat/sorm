package orm

import (
	"testing"
)

func Test_NewNid(t *testing.T) {
	t.Log("random nid:", NewNid())
	t.Log("random nid(time1):", NewNidTime())
	t.Log("random nid(time2):", NewNidTime())
}
