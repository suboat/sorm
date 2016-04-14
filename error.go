package orm

import (
	"errors"
)

var (
	ErrParamsType             error = errors.New("Params type error")
	ErrTableNotFound          error = errors.New("Not found table")
	ErrUnSupportedType        error = errors.New("Unsupported type error")
	ErrNotExist               error = errors.New("Not exist error")
	ErrCacheFailed            error = errors.New("Cache failed")
	ErrNeedDeletedCond        error = errors.New("Delete need at least one condition")
	ErrNotImplemented         error = errors.New("Not implemented")
	ErrDeleteObjectEmpty      error = errors.New("Delete Object empty")
	ErrUpdateObjectEmpty      error = errors.New("Update Object empty")
	ErrUpdateOneObjectMult    error = errors.New("Update One Object, but match mult")
	ErrFetchObjectEmpty       error = errors.New("Fetch object is empty")
	ErrFetchOneDuplicate      error = errors.New("Want to fetch one record, but duplicated")
	ErrUidInvalid             error = errors.New("Uid invalid")
	ErrUidEmpty               error = errors.New("Uid empty")
	ErrUidEmptyOrGuest        error = errors.New("Uid empty or is guest")
	ErrNidInvalid             error = errors.New("Nid invalid")
	ErrNidEmpty               error = errors.New("Nid empty")
	ErrMInvalid               error = errors.New("M invalid")
	ErrAccessionInvalid       error = errors.New("Accession invalid")
	ErrReflectValSet          error = errors.New("reflect value cat not set")
	ErrIndexStruct            error = errors.New("EnsureIndexWithTag: input is not a struct")
	ErrTransEmpty             error = errors.New("Transaction: empty")
	ErrTransRollbackUndefined error = errors.New("option: rollback error undefined")
	ErrModelUndefined         error = errors.New("model undefined")
)
