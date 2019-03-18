package mongo

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"
	"github.com/suboat/sorm/songo"

	//"gopkg.in/mgo.v2"
	"github.com/globalsign/mgo"

	"encoding/json"
	"fmt"
	"strings"
)

var (
	MaxOpenConns = 50    // 默认最大链接数
	CfgDbUnsafe  = false // true: fetch fields unsafe
	CfgTxUnsafe  = false // true: support trans can not rollback
)

// 数据库链接参数
type ArgConn struct {
	DbUrl  string `json:"dbUrl"`  //
	DbName string `json:"dbName"` //
}

// 数据库
type DatabaseSQL struct {
	ArgConn *ArgConn      //
	DB      *mgo.Database // 数据库
	Session *mgo.Session  // mongo session
	log     orm.Logger
}

//
func (db *DatabaseSQL) String() string {
	return orm.DriverNameMongo
}

// Version 打印数据库及版本号
func (db *DatabaseSQL) DriverName() string {
	return orm.DriverNameMongo
}

// 连接参数序列化
func (arg *ArgConn) String() (s string) {
	s = fmt.Sprintf("%s %s", arg.DbUrl, arg.DbName)
	return
}
func (arg *ArgConn) Json() (s string) {
	if b, _err := json.Marshal(arg); _err == nil {
		s = string(b)
	} else {
		s = "{}"
	}
	return
}

// 初始或重设数据库连接
func (db *DatabaseSQL) Reset() (err error) {
	// close && connect
	if db.Session != nil {
		db.Close()
	}
	if db.Session, err = mgo.Dial(db.ArgConn.DbUrl + db.ArgConn.DbName); err != nil {
		return
	} else {
		// Optional. Switch the session to a monotonic behavior.
		db.Session.SetMode(mgo.Monotonic, true)
	}
	db.DB = db.Session.DB(db.ArgConn.DbName)
	return
}

// 断开数据库连接
func (db *DatabaseSQL) Close() (err error) {
	if db.Session != nil {
		db.Session.Close()
	}
	return
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
	m.Collection = db.DB.C(s)
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

// ** other
// 解析参数
func argParser(jsonStr string) (db *DatabaseSQL, err error) {
	if len(jsonStr) == 0 {
		err = orm.ErrDbParamsEmpty
		return
	}
	db = new(DatabaseSQL)
	db.ArgConn = new(ArgConn)
	// 解析json参数
	arg := map[string]string{}
	_ = json.Unmarshal([]byte(jsonStr), &arg)
	// 兼容sql的参数读取
	db.ArgConn.DbUrl = arg["url"]
	if len(db.ArgConn.DbUrl) == 0 {
		db.ArgConn.DbUrl = arg["host"]
	}
	db.ArgConn.DbName = arg["db"]
	if len(db.ArgConn.DbName) == 0 {
		db.ArgConn.DbName = arg["dbname"]
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
	return
}

// 新键连接
func NewDb(arg string) (orm.Database, error) {
	var (
		db  *DatabaseSQL
		err error
	)
	if db, err = argParser(arg); err != nil {
		return nil, err
	}
	// 数据库初始化
	if err = db.Reset(); err != nil {
		return nil, err
	}
	return db, err
}

// register
func init() {
	orm.RegisterDriver(orm.DriverNameMongo, NewDb)
	// 用songo作为解析驱动
	orm.HookParseSafe = songo.ParseSafe //
	orm.HookParseMgo = songo.ParseMgo   // test
	// default log
	if orm.Log == nil {
		orm.SetLog(log.Log)
	}
}
