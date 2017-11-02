package main

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"

	// import drivers
	_ "github.com/suboat/sorm/pg"

	"time"
)

// Config
type Config struct {
	// config of postgres
	PgHost string
	PgPort string
	PgUser string
	PgPsw  string
	PgDb   string
}

// inherit example
type DemoPerson struct {
	Uid        orm.Uid          `sorm:"size(36) unique"`
	DemoAnimal `bson:",inline"` // mgo default embed fields

	FirstName string    `sorm:"size(32) index"`
	LastName  string    `sorm:"size(64) index"`
	Age       int       `sorm:"index"`
	Birthday  time.Time `sorm:"index"`
	Message   string
	Address   demoAddress
	Password  []byte
}

type DemoAnimal struct {
	Height float32 `sorm:"index"`
	Weight float32
	Length int64
}

type demoAddress struct {
	Road   string
	Number int
}

// example
func main() {
	var (
		cfg = &Config{
			PgHost: "192.168.8.138",
			PgPort: "54326",
			PgUser: "example",
			PgPsw:  "example",
			PgDb:   "example",
		}
		db  orm.Database
		tb orm.Model
		err error
	)

	connStr := `{"user":"` + cfg.PgUser + `", "password": "` + cfg.PgPsw + `", "host": "` +
		cfg.PgHost + `", "port": "` + cfg.PgPort + `", "dbname": "` + cfg.PgDb + `", "sslmode": "disable"}`

	if db, err = orm.New("pg", connStr); err != nil {
		panic(err)
	} else {
		log.Debug("connect success: ", db.String())
	}

	// sync table and fields with struct
	tb = db.Model("demoperson")
	if err = tb.EnsureIndexWithTag(&DemoPerson{});err != nil {
		panic(err)
	}

}
