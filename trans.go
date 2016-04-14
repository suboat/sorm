package orm

import (
	"database/sql"
)

// 事务transactions
type Trans interface {
	Commit() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error) // sqlx
	Rollback() error
	Error() error                                // 打印事务中积累的错误
	ErrorSet(error)                              // 人为设置事务错误,使AutoTrans触发Rollback
	Promise() []func(error)                      // func(error) 别的事件,error!=nil时会回滚
	PromiseBind(bind ...func(error)) (err error) // 绑定commit和rollback时出发的函数
}
