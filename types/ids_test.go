package types

import (
	"testing"
)

//
func Test_Accession(t *testing.T) {
	s := NewAccession()
	if err := IsAccessionValid(s); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("accession: %s, len: %d", s, len(s))
	}
}

//
func Test_NewNid(t *testing.T) {
	t.Logf("random nid: %s", NewNid())
	t.Logf("random nid(time1): %s", NewNidTime())
	t.Logf("random nid(time2): %s", NewNidTime())
	if err := IsNidValid(NewNidTime()); err != nil {
		t.Fatal(err)
	}
}

//
func Test_Uid(t *testing.T) {
	s := NewUID()
	if err := IsUIDValid(s); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("uid: %s, len: %d", s, len(s))
	}
}
