package postgres

import (
	"database/sql"

	"git.yichui.net/open/orm"
	"github.com/jmoiron/sqlx"
)

type Trans struct {
	Tx      *sqlx.Tx
	TxError error
	// promise
	promise []func(error)
}

func (t *Trans) Error() error {
	return t.TxError
}

func (t *Trans) ErrorSet(err error) {
	t.TxError = err
}

func (t *Trans) Commit() (err error) {
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

func (t *Trans) Rollback() error {
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
	return t.Tx.Rollback()
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
