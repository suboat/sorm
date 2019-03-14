# 连接数据库

支持的数据库:

- [x] PostgreSQL(9.x及以上)
- [x] MySQL(5.3以上)
- [x] MariaDB
- [x] MongoDB
- [ ] SQLite

## 快速入门

加载所需的数据库驱动:

```go
package main

import (
	"github.com/tudyzhb/go-sorm"
	"github.com/tudyzhb/go-sorm/log"

	_ "github.com/tudyzhb/go-sorm/driver/mongo"
	_ "github.com/tudyzhb/go-sorm/driver/mysql"
	_ "github.com/tudyzhb/go-sorm/driver/pg"
)

func main() {

	var (
		// config
		configPg    = `{"user":"suboat", "password": "suboat123", "host": "127.0.0.1", "port": "5432", "dbname": "suboat", "sslmode": "disable"}`
		configMysql = `{"user":"suboat", "password": "suboat123", "host": "127.0.0.1", "port": "3306", "dbname": "suboat", "sslmode": "disable"}`
		configMongo = `{"url":"mongodb://127.0.0.1:27017/", "db": "mgotest"}`
		// database
		dbPg    orm.Database
		dbMysql orm.Database
		dbMongo orm.Database
		//
		err error
	)

	// pg
	if dbPg, err = orm.New(orm.DbNamePostgres, configPg); err != nil {
		log.Error(err)
	} else {
		log.Debug(dbPg.String())
	}

	// mysql
	if dbMysql, err = orm.New(orm.DbNameMysql, configMysql); err != nil {
		log.Error(err)
	} else {
		log.Debug(dbMysql.String())
	}

	// mongo
	if dbMongo, err = orm.New(orm.DbNameMongo, configMongo); err != nil {
		log.Error(err)
	} else {
		log.Debug(dbMongo.String())
	}
}

```