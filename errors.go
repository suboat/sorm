// Code generated by i18n error platform, please DO NOT EDIT.
package orm

import (
	"errors"
)

var (
	// database
	ErrDbParamsInvalid error = errors.New("new database parms invalid") // 新建数据库参数有误
	ErrDbParamsEmpty   error = errors.New("new database parms empty")   // 新建数据库参数有误
	// hook
	ErrHookFuncUndefined error = errors.New("hook function undefined") // 函数未定义
	// sync
	ErrSyncEmbedPointNil error = errors.New("sync embed field ponitor nil") // 内嵌指针结构未初始化
	// model:trans
	ErrTransNotSupport        error = errors.New("driver not support trans")                // 驱动不支持事物: 如mongodb
	ErrTransNotSupportMethod  error = errors.New("driver not support this method of trans") // 驱动不支持事物: 如mongodb
	ErrTransEmpty             error = errors.New("params transaction empty")                // 事物作为参数是空
	ErrTransInvalid           error = errors.New("params transaction invalid")              // 事物作为参数非法
	ErrTransRollbackUndefined error = errors.New("option: rollback-error undefined")        // 事物要回滚，但未指明错误
	ErrTransLockWholeTable    error = errors.New("trans lock whole table")                  // 没有where语句的lock，不允许
	ErrTransLevelUnknown      error = errors.New("trans level unknown")                     // 事物级别未知
	// query
	ErrMatchNone     error = errors.New("match none")     // 无匹配记录
	ErrMatchExist    error = errors.New("match exist")    // 记录已存在
	ErrMatchMultiple error = errors.New("match multiple") // 期望搜到一条记录，但是返回多条
	// update
	ErrUpdateMapKeyInvalid   error = errors.New("update parms map-key invalid")  // 更新结构的key非法
	ErrUpdateMapTypeUnknown  error = errors.New("update parms map type unknown") // 更新的输入参数不支持
	ErrUpdateIncValueInvalid error = errors.New("update parms inc-val invalid")  // 更新结构的value非法
	// index
	ErrIndexTextParamsInvalid error = errors.New("text index params error") // 全文索引参数错误
)