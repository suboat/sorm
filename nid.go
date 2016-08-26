package orm

import (
	"database/sql/driver"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// 纯数字的12位ID, 前6位表示日期, 后6位表示流水号
type Nid string

const (
	EmptyNid Nid = "100000000000"
	NidLen       = len(EmptyNid)
)

var (
	SeedsNum = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	NidSeed  = int(rand.New(rand.NewSource(time.Now().UnixNano())).Float64() * 1000000) // 随机起始数
	NidMux   = new(sync.Mutex)                                                          // 种子锁
)

func (u Nid) String() string {
	return string(u)
}

func (u Nid) Valid() (err error) {
	if len(u.String()) == 0 {
		err = ErrNidInvalid
	}
	return
}

// 是否为空
func (u Nid) IsEmpty() bool {
	return len(u) == 0 || u == EmptyNid
}

// sql: Value implements the driver.Valuer interface
func (u Nid) Value() (driver.Value, error) {
	return string(u), nil
}

// sql: Scan implements the sql.Scanner interface
func (u *Nid) Scan(src interface{}) error {
	*u = Nid(string(src.([]uint8)))
	return nil
}

func NewNid() (n Nid) {
	n = EmptyNid
	for n == EmptyNid {
		var s [NidLen]string
		for i, _ := range s {
			if i == 0 {
				s[i] = SeedsNum[1+rand.Intn(9)]
			} else {
				s[i] = SeedsNum[rand.Intn(10)]
			}
		}
		n = Nid(s[0] + s[1] + s[2] + s[3] + s[4] + s[5] + s[6] + s[7] + s[8] + s[9] + s[10] + s[11])
	}
	return
}

func NewNidTime() (n Nid) {
	n = EmptyNid
	NidMux.Lock()
	defer NidMux.Unlock()

	if NidSeed = NidSeed + 1; NidSeed > 999999 {
		NidSeed = 1
	}

	t := time.Now()
	//n = Nid(fmt.Sprintf("%d%02d%02d%04d", t.Year(), t.Month(), t.Day(), NidSeed))
	n = Nid(fmt.Sprintf("%s%06d", t.Format("060102"), NidSeed))
	return
}
