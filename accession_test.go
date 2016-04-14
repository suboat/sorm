package orm

import (
	"testing"
)

func Test_Accession(t *testing.T) {
	s := NewAccession()
	t.Log("accession ", s, "len:", len(s))
}
