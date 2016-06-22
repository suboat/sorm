package orm

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/satori/go.uuid"
)

const (
	GuestUid Uid = "11111111-1111-1111-1111-111111111111" // 匿名用户的uid
)

type Uid string
type UidLis []Uid

func (u Uid) String() string {
	return string(u)
}

func (u Uid) Valid() (err error) {
	if len(u.String()) == 0 {
		err = ErrUidInvalid
	}
	return
}

// 是否为空
func (u Uid) IsEmpty() bool {
	return len(u) == 0
}

// 是否是匿名
func (u Uid) IsGuest() bool {
	return u == GuestUid
}

// 要求不是空或游客
func (u Uid) AssertValid() (err error) {
	if u.IsEmpty() || u.IsGuest() {
		err = ErrUidEmptyOrGuest
	}
	return
}

// sql: Value implements the driver.Valuer interface
func (u Uid) Value() (driver.Value, error) {
	return string(u), nil
}

// sql: Scan implements the sql.Scanner interface
func (u *Uid) Scan(src interface{}) error {
	if _s, _ok := src.(string); _ok == true {
		*u = Uid(_s)
	} else {
		*u = Uid(string(src.([]uint8)))
	}
	return nil
}

// uid lis
func (ul UidLis) Value() (driver.Value, error) {
	return json.Marshal(ul)
}
func (ul *UidLis) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), ul)
}

func NewUid() Uid {
	return Uid(uuid.NewV4().String())
}
