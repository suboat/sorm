package types

import (
	"github.com/stretchr/testify/require"

	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

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
func Test_Accession12(t *testing.T) {
	as := require.New(t)
	for i := 0; i < 300000; i++ {
		id := NewAccession12()
		id = strings.Replace(id, "-", "", -1)
		as.Equal("00000000", id[0:8])
		b, _ := hex.DecodeString(id[8:])
		c := base64.RawURLEncoding.EncodeToString(b)
		b2, _ := base64.RawURLEncoding.DecodeString(c)
		id2 := fmt.Sprintf("00000000%x", b2)
		// debug
		//t.Logf("%s b:%d c:%d %s %s", c, len(b), len(c), id, id2)
		// assert
		as.Equal(id2, id)
		as.Equal(true, len(b) <= 12)
		as.Equal(true, len(c) <= 16)
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
