package mysql

import (
	"github.com/stretchr/testify/require"

	"testing"
)

//
func Test_TimeZone(t *testing.T) {
	as := require.New(t)
	timeStr := "2019-05-21T14:38:01+08:00"
	target := "2019-05-21 06:38:01"
	as.Equal(true, RegTimeWithZone.MatchString(timeStr))
	as.Equal(false, RegTimeWithZone.MatchString(target))
	as.Equal(target, PubTimeConvert(timeStr))
}
