package mysql

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"
	"github.com/suboat/sorm/songo"

	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql" // 驱动包
	"github.com/jmoiron/sqlx"
)

var (
	// MaxOpenConns 默认最大链接数
	MaxOpenConns = 50
	// CfgDbUnsafe false:数据库严格映射到结构体
	CfgDbUnsafe = false // true: sqlx.Unsafe 防止报错 https://github.com/jmoiron/sqlx/blob/master/sqlx.go#L601
	// 针对数据库
	driverName = orm.DriverNameMysql //
)

const (
	// DbVerMysql mysql版本
	DbVerMysql = "mysql"
	// DbVerMaria mariadb版本
	DbVerMaria = "maria"
)

// ArgConn 数据库链接参数
type ArgConn struct {
	Driver string     `json:"driver"` //
	Params url.Values `json:"params"` //
	//
	Host     string `json:"host"`     //
	Port     string `json:"port"`     //
	Database string `json:"database"` //
	User     string `json:"user"`     //
	Password string `json:"password"` //
}

// DatabaseSQL 数据库
type DatabaseSQL struct {
	ArgConn *ArgConn
	DB      *sqlx.DB
	log     orm.Logger
}

//
func (arg *ArgConn) Init() (err error) {
	switch arg.Driver {
	case orm.DriverNamePostgres:
		if arg.Params == nil {
			arg.Params = url.Values{}
			arg.Params.Add("sslmode", "disable")
		}
		if len(arg.Port) == 0 {
			arg.Port = "5432"
		}
		if len(arg.Host) == 0 {
			arg.Host = "127.0.0.1"
		}
	case orm.DriverNameMysql:
		if arg.Params == nil {
			arg.Params = url.Values{}
			arg.Params.Add("parseTime", "true")
			arg.Params.Add("collation", "utf8mb4_general_ci") // for emoji
		}
		if len(arg.Port) == 0 {
			arg.Port = "3306"
		}
		if len(arg.Host) == 0 {
			arg.Host = "127.0.0.1"
		}
	case orm.DriverNameSQLite:
		if len(arg.Database) == 0 {
			arg.Database = "sqlite.db"
		}
		break
	case orm.DriverNameMongo:
		break
	default:
		return orm.ErrDbParamsInvalid
	}
	return
}

// 连接参数序列化
func (arg *ArgConn) String() (s string) {
	_ = arg.Init()
	switch arg.Driver {
	case orm.DriverNamePostgres:
		s = fmt.Sprintf("%s://%s:%s@%s:%s/%s?%s",
			arg.Driver, arg.User, arg.Password, arg.Host, arg.Port, arg.Database, arg.Params.Encode())
	case orm.DriverNameMysql:
		s = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
			arg.User, arg.Password, arg.Host, arg.Port, arg.Database, arg.Params.Encode())
	case orm.DriverNameSQLite:
		s = arg.Database
	default:
		panic("unknown database core")
	}
	return
}

//
func (arg *ArgConn) ParseFromJSON(jsonStr string) (err error) {
	if arg == nil {
		return orm.ErrDbParamsInvalid
	}
	if len(jsonStr) == 0 {
		err = orm.ErrDbParamsEmpty
		return
	}
	if err = json.Unmarshal([]byte(jsonStr), arg); err != nil {
		return
	}
	return arg.Init()
}

//
func (db *DatabaseSQL) String() string {
	return db.DB.DriverName()
}

// DriverName 数据库驱动名称
func (db *DatabaseSQL) DriverName() string {
	return db.DB.DriverName()
}

// Reset 初始或重设数据库连接
func (db *DatabaseSQL) Reset() (err error) {
	// close && connect
	if err = db.Close(); err != nil {
		return
	}
	if db.DB, err = sqlx.Connect(db.ArgConn.Driver, db.ArgConn.String()); err != nil {
		return
	} else if CfgDbUnsafe {
		db.DB = db.DB.Unsafe()
	}
	// version
	db.log.Infof("[conn] %s connected", db.String())
	return
}

// Close 断开数据库连接
func (db *DatabaseSQL) Close() (err error) {
	if db.DB == nil {
		return
	}
	return db.DB.Close()
}

// Model 获取table
func (db *DatabaseSQL) Model(s string) orm.Model {
	return db.ModelWith(s, nil)
}

// ModelWith 获取table
func (db *DatabaseSQL) ModelWith(s string, arg *orm.ArgModel) orm.Model {
	s = strings.Replace(strings.ToLower(s), `"`, ``, -1)
	m := new(Model)
	m.TableName = s
	m.DatabaseSQL = db
	if db.log != nil {
		if _log, ok := db.log.(*log.Logger); ok {
			m.log = _log.Copy()
		}
	}
	if m.log == nil {
		m.log = log.Log.Copy()
	}
	if arg != nil && arg.LogLevel > 0 && m.log != nil {
		m.log.SetLevel(arg.LogLevel)
	}
	return m
}

// NewDb 新键连接
func NewDb(arg string) (ret orm.Database, err error) {
	var (
		con = &ArgConn{Driver: driverName}
		db  = new(DatabaseSQL)
	)
	if err = con.ParseFromJSON(arg); err != nil {
		return
	} else {
		db.ArgConn = con
	}
	// log
	if orm.Log != nil {
		if _log, ok := orm.Log.(*log.Logger); ok {
			db.log = _log.Copy()
		}
	}
	if db.log == nil {
		db.log = log.Log.Copy()
	}
	//
	if err = db.Reset(); err != nil {
		return nil, err
	}
	// 设置重要参数
	db.DB.SetMaxOpenConns(MaxOpenConns)
	//
	ret = db
	return
}

// register
func init() {
	orm.RegisterDriver(driverName, NewDb)
	// 用songo作为解析驱动
	orm.HookParseSafe = songo.ParseSafe             //
	orm.HookParseSQL[driverName] = songo.ParseMysql //
	// default log
	if orm.Log == nil {
		orm.SetLog(log.Log)
	}
}
