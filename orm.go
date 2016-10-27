package orm

import (
	"fmt"
	"math/rand"
	"time"
)

type Hash uint

const (
	Mongo    Hash = 1 + iota // orm/mongo
	Mysql                    // 预留
	Postgres                 // postgresql
	maxHash
)

// Tag key
const (
	TagAccession = "accession"
	TagUid       = "uid"
	TagKey       = "sorm"
	// mean value
	TagValNo   = `%no%`   // 不等于
	TagValNe   = `%ne%`   // 不等于 alias
	TagValLike = `%like%` // like
	TagValLt   = `%lt%`   // 小于
	TagValLte  = `%lte%`  // 小于等于
	TagValGt   = `%gt%`   // 大于
	TagValGte  = `%gte%`  // 大于等于
	// mean key
	TagQueryKeyOr  = `%or%`
	TagQueryKeyAnd = `%and%`
	TagQueryKeyIn  = `%in%`
	// mean update
	TagUpdateInc = `%inc%` // 数据库级别增
)

var (
	// log
	Debug          = false
	DebugLevel int = 0
	// trans
	//TransTimeout = time.Second * 20 // debug
	TransTimeout = time.Minute * 10 // 事务悬浮10分钟后自动回滚
)

type Database interface {
	String() string     // 可打印
	Model(string) Model // 获取table或者collection
	Close() error       // 可关闭
}

var hashes = make([]func(string) (Database, error), maxHash)

// RegisterHash registers a function that returns a new instance of the given
// hash function. This is intended to be called from the init function in
// packages that implement hash functions.
func RegisterHash(h Hash, f func(string) (Database, error)) {
	if h >= maxHash {
		panic("orm: RegisterHash of unknown hash function")
	}
	hashes[h] = f
}

// 返回可操作数据库
func New(s string, arg string) (db Database, err error) {
	err = ErrParamsType
	if s == "mongo" {
		db, err = hashes[Mongo](arg)
	} else if s == "pg" {
		db, err = hashes[Postgres](arg)
	}
	return
}

func init() {
	//println("orm mongo")
	rand.Seed(time.Now().UnixNano())

	// ZoneOffset on server
	if _, offset := time.Now().Zone(); offset > 0 {
		ZoneOffset = fmt.Sprintf("+%02d:00", offset/60/60)
	} else {
		ZoneOffset = fmt.Sprintf("-%02d:00", offset/60/60)
	}
}
