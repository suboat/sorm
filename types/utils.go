package types

import (
	"github.com/satori/go.uuid"
)

// NewUUID 新UUID
func NewUUID() (ret string) {
	return uuid.NewV4().String()
}
