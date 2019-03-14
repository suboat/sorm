package orm

import (
	"database/sql"
)

const (
	// TransReadUncommitted 读未提交:最低级别，任何情况都无法保证 隐患:脏读
	TransReadUncommitted = "Read uncommitted"
	// TransReadCommitted 读提交:可避免脏读的发生 隐患:事务并发读
	TransReadCommitted = "Read committed"
	// TransRepeatableRead 重复读:可避免脏读、不可重复读的发生 隐患:并发写入
	TransRepeatableRead = "Repeatable read"
	// TransSerializable 串行化:可避免脏读、不可重复读、幻读的发生
	TransSerializable = "Serializable"
)

// Trans 事务
type Trans interface {
	//
	Commit() error   // 提交事务,多次提交不报错
	Rollback() error // 回滚事务,多次提交不报错
	//
	Error() error                              // 打印事务中积累的错误
	ErrorSet(error)                            // 人为设置事务错误,使AutoTrans触发Rollback
	Promise() []func(error)                    // func(error) 别的事件, error!=nil时会回滚
	PromiseAdd(pfn ...func(error)) (err error) // 绑定commit和rollback时触发的函数
	//
	Exec(query string, args ...interface{}) (sql.Result, error) //

	//DebugPush(info ...string) error            // DEBUG: 往事务中记录信息,方便出错时打印调试
	// sqlx的注意方法
	//NamedExec(query string, arg interface{}) (sql.Result, error)      // sqlx
	//Get(dest interface{}, query string, args ...interface{}) error    // 计数
	//Select(dest interface{}, query string, args ...interface{}) error // 搜索
}
