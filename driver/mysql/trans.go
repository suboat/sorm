package mysql

import (
	"github.com/jmoiron/sqlx"
	"github.com/suboat/sorm"

	"database/sql"
	"strings"
	"time"
)

// Trans 事务实现
type Trans struct {
	Tx      *sqlx.Tx
	TxError error
	// promise
	promise []func(error)
	// already commit or rollback
	isFinish bool
	// sql history, for debug
	debugInfo []string
	// trans expired
	timer *time.Timer
}

func (t *Trans) Error() error {
	return t.TxError
}

// ErrorSet 设置错误
func (t *Trans) ErrorSet(err error) {
	t.TxError = err
}

// Commit 事务提交
func (t *Trans) Commit() (err error) {
	if t.isFinish {
		return
	}
	if t.TxError != nil {
		if err = t.Tx.Rollback(); err != nil {
			return
		}
		err = t.TxError
	} else if err = t.Tx.Commit(); err != nil {
		t.TxError = err
	}

	// promise
	if t.promise != nil {
		for _, fn := range t.promise {
			fn(err)
		}
	}

	t.isFinish = true
	t.timerReset()
	return
}

// Rollback 回滚事务
func (t *Trans) Rollback() (err error) {
	if t.isFinish {
		return
	}

	// ignore t.TxError

	// promise
	if t.promise != nil {
		var pErr error
		if pErr = t.TxError; pErr == nil {
			pErr = orm.ErrTransRollbackUndefined
		}
		for _, fn := range t.promise {
			fn(pErr)
		}
	}
	err = t.Tx.Rollback()

	t.isFinish = true
	t.timerReset()
	return
}

// Promise 返回已绑定
func (t *Trans) Promise() []func(error) {
	return t.promise
}

// PromiseAdd 添加绑定
func (t *Trans) PromiseAdd(fns ...func(error)) (err error) {
	if t.promise == nil {
		t.promise = []func(error){}
	}
	for _, fn := range fns {
		if fn != nil {
			t.promise = append(t.promise, fn)
		}
	}
	return
}

// Exec 执行
func (t *Trans) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	if t.TxError != nil {
		err = t.TxError
		return
	}
	if result, err = t.Tx.Exec(query, args...); err != nil {
		t.TxError = err
	}
	return
}

// Get 在事务中获取
func (t *Trans) Get(dest interface{}, query string, args ...interface{}) (err error) {
	if t.TxError != nil {
		err = t.TxError
		return
	}
	if err = t.Tx.Unsafe().Get(dest, query, args...); err != nil {
		t.TxError = err
	}
	return
}

// Select 在事务中查询
func (t *Trans) Select(dest interface{}, query string, args ...interface{}) (err error) {
	if t.TxError != nil {
		err = t.TxError
		return
	}
	if err = t.Tx.Select(dest, query, args...); err != nil {
		t.TxError = err
	}
	return
}

// NamedExec 用结构体字段依赖执行
func (t *Trans) NamedExec(query string, arg interface{}) (result sql.Result, err error) {
	if t.TxError != nil {
		err = t.TxError
		return
	}
	if result, err = t.Tx.NamedExec(query, arg); err != nil {
		t.TxError = err
	}
	return
}

// DebugPush 记录调试信息
func (t *Trans) DebugPush(info ...string) (err error) {
	t.debugInfo = append(t.debugInfo, info...)
	return
}

func (t *Trans) debugReport() (s string) {
	if t.debugInfo == nil {
		return
	}
	s = strings.Join(t.debugInfo, ",")
	return
}

func (t *Trans) timerReset() {
	if t.timer != nil && t.isFinish {
		t.timer.Reset(0 * time.Second)
	}
}
