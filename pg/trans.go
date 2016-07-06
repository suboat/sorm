package postgres

import (
	"database/sql"

	"github.com/suboat/sorm"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

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

func (t *Trans) ErrorSet(err error) {
	t.TxError = err
}

func (t *Trans) DebugPush(info ...string) (err error) {
	if t.debugInfo == nil {
		t.debugInfo = []string{}
	}
	for _, inf := range info {
		t.debugInfo = append(t.debugInfo, inf)
	}
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
	if t.timer != nil && t.isFinish == true {
		t.timer.Reset(0 * time.Second)
	}
}

func (t *Trans) Commit() (err error) {
	if t.isFinish == true {
		return
	}
	if t.TxError != nil {
		t.Tx.Rollback()
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

func (t *Trans) Rollback() (err error) {
	if t.isFinish == true {
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

func (t *Trans) Promise() []func(error) {
	return t.promise
}

func (t *Trans) PromiseBind(fns ...func(error)) (err error) {
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
