package orm

import (
	"database/sql/driver"
	"github.com/satori/go.uuid"
)

// 外部访问id的建议模型
type Accession string

// 是否合法
func (a Accession) Valid() error {
	if len(string(a)) > 0 {
		return nil
	} else {
		return ErrAccessionInvalid
	}
}

// 是否为空
func (a Accession) IsEmpty() bool {
	return len(a) == 0
}

// string
func (a Accession) String() string {
	return string(a)
}

// sql: Value implements the driver.Valuer interface
func (a Accession) Value() (driver.Value, error) {
	return string(a), nil
}

// sql: Scan implements the sql.Scanner interface
func (a *Accession) Scan(src interface{}) error {
	if _s, _ok := src.(string); _ok == true {
		*a = Accession(_s)
	} else {
		*a = Accession(string(src.([]uint8)))
	}
	return nil
}

func NewAccession() Accession {
	return Accession(uuid.NewV4().String())
}
