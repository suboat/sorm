package orm

import (
	"database/sql"
)

const (
	TransReadUncommitted = "Read uncommitted"
	TransReadCommitted   = "Read committed"
	TransRepeatableRead  = "Repeatable read"
	TransSerializable    = "Serializable"
)

// 事务transactions
type Trans interface {
	Commit() error                                                    // 提交事务,多次提交不报错
	Exec(query string, args ...interface{}) (sql.Result, error)       //
	NamedExec(query string, arg interface{}) (sql.Result, error)      // sqlx
	Get(dest interface{}, query string, args ...interface{}) error    // 计数
	Select(dest interface{}, query string, args ...interface{}) error // 搜索
	Rollback() error                                                  // 回滚事务,多次提交不报错
	Error() error                                                     // 打印事务中积累的错误
	ErrorSet(error)                                                   // 人为设置事务错误,使AutoTrans触发Rollback
	Promise() []func(error)                                           // func(error) 别的事件,error!=nil时会回滚
	PromiseBind(bind ...func(error)) (err error)                      // 绑定commit和rollback时出发的函数
	DebugPush(info ...string) error                                   // DEBUG: 往事务中记录信息,方便出错时打印调试
}
