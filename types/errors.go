// Code generated by i18n error platform, please DO NOT EDIT.
package types

import (
	"errors"
)

var (
	// ids
	ErrUidInvalid      error = errors.New("uid invalid")           // uid:为游客或系统
	ErrUidEmpty        error = errors.New("uid empty")             // uid:空
	ErrUidEmptyOrGuest error = errors.New("uid empty or is guest") // uid:
	ErrNidInvalid      error = errors.New("nid invalid")           // nid:非法
	ErrNidEmpty        error = errors.New("nid empty")             // nid:空
)