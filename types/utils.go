package types

import (
	"encoding/binary"
	"github.com/satori/go.uuid"

	"time"
)

// NewUUID 新UUID
func NewUUID() (ret string) {
	return uuid.NewV4().String()
}

// NewUnix 前4字节为时间戳的uuid
func NewUnix() (ret string) {
	d := uuid.NewV4()
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(time.Now().Unix()))
	d[0] = b[4]
	d[1] = b[5]
	d[2] = b[6]
	d[3] = b[7]
	return d.String()
}
