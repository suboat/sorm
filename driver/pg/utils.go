package pg

import (
	"fmt"
	"github.com/suboat/sorm"
	"strings"
)

//
func PubFieldWrap(s string) (ret string) {
	return fmt.Sprintf(`"%s"`, strings.ToLower(s))
}

//
func PubFieldWrapAll(s []string) (ret []string) {
	for _, v := range s {
		ret = append(ret, PubFieldWrap(v))
	}
	return
}

//
func PubFieldWrapByFieldInfo(s []*orm.FieldInfo) (ret []string) {
	for _, v := range s {
		ret = append(ret, PubFieldWrap(v.Name))
	}
	return
}

//
func PubFieldWrapByDest(dest interface{}) (ret []string) {
	if _ret, _err := orm.StructModelInfoByDest(dest); _err == nil {
		return PubFieldWrapByFieldInfo(_ret)
	}
	return
}
