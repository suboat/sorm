package songo

import (
	"github.com/stretchr/testify/require"
	"testing"
)

//
func Test_SafeField(t *testing.T) {
	as := require.New(t)
	as.Nil(nil)
	as.Equal("tree", SafeField(`Tree`))
	as.Equal("tree.leaf", SafeField(`Tree.Leaf`))
	as.Equal("tree_table.leaf-one", SafeField(`Tree_table.Leaf-One`))
	as.Equal("", SafeField(`%`))
	as.Equal("", SafeField(`*`))
	as.Equal("", SafeField("`"))
	as.Equal("", SafeField(`'`))
	as.Equal("", SafeField(`"`))
}
