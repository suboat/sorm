package types

import (
	"github.com/satori/go.uuid"

	cryptoRand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// **** uid
const (
	UIDSystem string = "00000000-0000-0000-0000-000000000000" // 系统uid
	UIDGuest  string = "11111111-1111-1111-1111-111111111111" // 匿名用户的uid
)

// IsUIDValid uid是否合法
func IsUIDValid(s string) (err error) {
	if len(s) == 0 {
		err = ErrUidEmpty
		return
	}
	if s == UIDGuest {
		err = ErrUidEmptyOrGuest
		return
	}
	if s == UIDSystem {
		err = ErrUidInvalid
		return
	}
	_, err = uuid.FromString(s)
	return
}

// NewUID 新uid
func NewUID() string {
	return NewUUID()
}

// IsAccessionValid accession是否合法
func IsAccessionValid(s string) (err error) {
	_, err = uuid.FromString(s)
	return
}

// NewAccession 新建acc
func NewAccession() string {
	return NewUUID()
}

// ** serial / nid
// 纯数字的12位ID, 前6位表示日期, 后6位表示流水号
const (
	EmptyNid string = "100000000000"
	NidLen          = len(EmptyNid)
)

var (
	// SeedsNum 随机数内容
	SeedsNum = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	// NidSeed 随机起始数
	NidSeed = int(rand.New(rand.NewSource(time.Now().UnixNano())).Float64() * 1000000)
	// NidMux 种子锁
	NidMux = new(sync.Mutex)
)

// IsNidValid 合法性判断
func IsNidValid(s string) (err error) {
	if len(s) == 0 {
		err = ErrNidEmpty
		return
	}
	if false {
		err = ErrNidInvalid
		return
	}
	return
}

// NewNid 新Nid
func NewNid() (n string) {
	n = EmptyNid
	for n == EmptyNid {
		var s [NidLen]string
		for i := range s {
			if i == 0 {
				s[i] = SeedsNum[1+rand.Intn(9)]
			} else {
				s[i] = SeedsNum[rand.Intn(10)]
			}
		}
		n = s[0] + s[1] + s[2] + s[3] + s[4] + s[5] + s[6] + s[7] + s[8] + s[9] + s[10] + s[11]
	}
	return
}

// NewNidTime 依据时间的Nid
func NewNidTime() string {
	NidMux.Lock()
	NidSeed++
	if NidSeed > 999999 {
		NidSeed = 1
	}
	NidMux.Unlock()
	return fmt.Sprintf("%s%06d", time.Now().Format("060102"), NidSeed)
}

// 设置随机数种子
func initIds() (err error) {
	var (
		seed int64
	)
	if err = binary.Read(cryptoRand.Reader, binary.LittleEndian, &seed); err != nil {
		return
	}
	rand.Seed(seed)
	return
}
