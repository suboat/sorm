package postgres

import (
	"errors"
)

var (
	ErrTransExecParams error = errors.New("trans exec params error")
	ErrUpdateMapTyep   error = errors.New("update map type error")
)
