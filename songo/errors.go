// Code generated by i18n error platform, please DO NOT EDIT.
package songo

import (
	"errors"
)

var (
	ErrSongoMapUndefined       error = errors.New("songo map undefined")    //
	ErrSongoFormatInvalid      error = errors.New("songo format invalid")   // 格式有误
	ErrSongoMapKeyInvalid      error = errors.New("songo map key invalid")  // key中有非法字符
	ErrSongoMapOperatorInvalid error = errors.New("songo operator invalid") // 比较符非法
	ErrSongoMapValMapMultiple  error = errors.New("songo val-map invalid")  // map表示的val只有一对key-val
	ErrSongoMapDeepOutOf       error = errors.New("songo map deep out of")  // map层太深
)
