package main

import (
	"github.com/suboat/sorm"

	// import drivers
	_ "github.com/suboat/sorm/pg"
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
		err error
	)

	connStr := `{"user":"` + cfg.PgUser + `", "password": "` + cfg.PgPsw + `", "host": "` +
		cfg.PgHost + `", "port": "` + cfg.PgPort + `", "dbname": "` + cfg.PgDb + `", "sslmode": "disable"}`

	if db, err = orm.New("pg", connStr); err != nil {
		panic(err)
	}

	println("connect success: ", db.String())
}
