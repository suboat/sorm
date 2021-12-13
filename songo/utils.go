package songo

import (
	"regexp"
	"strings"
)

var (
	// 字段取值
	RegValTypeField = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
)

// safe sql fragment of field
func SafeField(s string) string {
	return RegValTypeField.FindString(strings.ToLower(s))
}
