package mysql

import (
	"github.com/stretchr/testify/require"
	"testing"
)

//
func Test_ArgConn(t *testing.T) {
	as := require.New(t)
	as.Nil(nil)
	arg := &ArgConn{Driver: driverName}
	t.Logf("arg %s", arg.String())
}
