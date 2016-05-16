package postgres

import (
	"encoding/json"
	"fmt"
	//"database/sql"

	"git.yichui.net/open/orm"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	DatabaseType = "postgres"
	DatabaseHash = orm.Postgres
)

// 数据库链接参数
type ArgConn struct {
	DbType   string
	DbName   string
	User     string
	Password string
	Host     string
	Post     string
	Port     string
	Sslmode  string // postgres
}

// 数据库
type DatabaseSql struct {
	ArgConn *ArgConn
	DB      *sqlx.DB
}

// print db
func (db *DatabaseSql) String() (s string) {
	s = db.ArgConn.DbName
	return
}

// 连接参数序列化
func (arg *ArgConn) String() (s string) {
	switch arg.DbType {
	case "postgres":
		s = fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s", arg.DbType, arg.User, arg.Password, arg.Host, arg.Port, arg.DbName, arg.Sslmode)
		break
	}
	return
}

// 初始或重设数据库连接
func (db *DatabaseSql) Reset() (err error) {
	// close && connect
	db.Close()
	if db.DB, err = sqlx.Connect(db.ArgConn.DbType, db.ArgConn.String()); err != nil {
		return
	}
	//db.DB = db.DB.Unsafe()
	return
}

// TODO:Close
func (db *DatabaseSql) Close() (err error) {
	return
}

// 获取table
func (db *DatabaseSql) Model(s string) orm.Model {
	m := new(Model)
	m.TableName = s
	m.DatabaseSql = db
	return m
}

// 解析参数
func argParser(jsonStr string) (db *DatabaseSql, err error) {
	if len(jsonStr) == 0 {
		err = orm.ErrParamsType
		return
	}
	db = new(DatabaseSql)
	arg := new(ArgConn)
	json.Unmarshal([]byte(jsonStr), &arg)
	arg.DbType = DatabaseType

	// TODO: database type
	arg.DbType = DatabaseType

	switch arg.DbType {
	case "postgres":
		if len(arg.Port) == 0 {
			arg.Port = "5432"
		}
		if len(arg.Sslmode) == 0 {
			arg.Sslmode = "require"
		}
		break
	}

	db.ArgConn = arg
	return
}

// new connect
func NewDb(arg string) (orm.Database, error) {
	var (
		db  *DatabaseSql
		err error
	)

	if db, err = argParser(arg); err != nil {
		return nil, err
	}
	if err = db.Reset(); err != nil {
		return nil, err
	}

	return db, err
}

// register
func init() {
	orm.RegisterHash(orm.Postgres, NewDb)
}
