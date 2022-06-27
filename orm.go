package orm

import (
	"time"
)

// 约定的命名
const (
	// 目前支持的数据库
	DriverNameMongo    = "mongo"
	DriverNameMysql    = "mysql"
	DriverNamePostgres = "postgres"
	DriverNameSQLite   = "sqlite3"
	DriverNameMsSql    = "mssql"
)

// project key
const (
	OrmKey        = "sorm"    //
	OrmTagJSON    = "json"    // string to json
	OrmTagChar    = "char"    // string to char
	OrmTagSize    = "size"    // string size
	OrmTagIndex   = "index"   //
	OrmTagUnique  = "unique"  //
	OrmTagPrimary = "primary" //
	OrmTagSerial  = "serial"  //
)

// Tag key
const (
	// mean value
	TagSep     = '$'      //
	TagValNe   = `$ne$`   // 不等于
	TagValNo   = `$no$`   // 不等于 alias
	TagValLt   = `$lt$`   // 小于
	TagValLte  = `$lte$`  // 小于等于
	TagValGt   = `$gt$`   // 大于
	TagValGte  = `$gte$`  // 大于等于
	TagValLike = `$like$` // like
	TagValText = `$text$` // 全文搜索
	// mean key
	TagQueryKeyOr  = `$or$`
	TagQueryKeyAnd = `$and$`
	TagQueryKeyIn  = `$in$`
	// mean update
	TagUpdateInc = `$inc$` // 数据库级别增
)

const (
	// LevelPanic level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	LevelPanic int = iota
	// LevelFatal level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	LevelFatal
	// LevelError level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	LevelError
	// LevelWarn level. Non-critical entries that deserve eyes.
	LevelWarn
	// LevelInfo level. General operational entries about what's going on inside the
	// application.
	LevelInfo
	// LevelDebug level. Usually only enabled when debugging. Very verbose logging.
	LevelDebug
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

// default
var (
	// Log 默认的日志输出
	Log Logger
	// LogLevel 目前全局日志级别
	LogLevel = LevelError
	// TransTimeout 事务悬浮10分钟后自动回滚
	TransTimeout = time.Minute * 10
	// HookParseSafe : 将map过滤为安全的map
	HookParseSafe = defaultHookParseSafe
	// HookParseSQL : 将map转为sql
	HookParseSQL = make(map[string]func(m map[string]interface{}, idx int) (sql string, vals []interface{}, err error))
	// HookParseMgo : 将map转为mgo
	HookParseMgo = defaultHookParseMgo
	// DefaultTimeStr 默认时间字符串
	DefaultTimeStr = "0001-01-01 00:00:00+00"
	// register
	hashes = make(map[string]func(string) (Database, error))
)

// Meta 基于数据库习惯的meta命名规则
type Meta struct {
	// 逻辑信息
	Ver string `json:"ver,omitempty"` // meta协议版本号
	// 内容
	Count   int `json:"count"`             // 同sql里的 count/total，总数
	Limit   int `json:"limit"`             // 同sql里的 limit
	Page    int `json:"page"`              // 0起始.直接告知当前页码,免去通过skip与limit计算
	SumJson int `json:"sumJson,omitempty"` // 返回聚合计算结果
	// 可选信息
	Skip  int         `json:"skip,omitempty"`  // 可选，不常用. 同sql里的 skip
	Num   int         `json:"num,omitempty"`   // 可选信息.直接告知此次返回的数据条数,免去读取数据字段
	Key   interface{} `json:"key,omitempty"`   // 可选信息.刚才用户提交的搜索信息
	Sort  interface{} `json:"sort,omitempty"`  // 可选信息.刚才用户提交的排序信息
	Group interface{} `json:"group,omitempty"` // 可选信息.刚才用户提交的聚合计算参数
	Sum   interface{} `json:"sum,omitempty"`   // 可选信息.刚才用户提交的聚合信息
}

// Logger 日志输出
type Logger interface {
	//
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	//
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	//
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	//
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	//
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
	//
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
	//
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	//
	SetLevel(level int)
	GetLevel() int
	Output(calldepth int, s string) error
}

// RegisterDriver registers a function that returns a new instance of the given
// hash function. This is intended to be called from the init function in
// packages that implement hash functions.
func RegisterDriver(driverName string, f func(string) (Database, error)) {
	hashes[driverName] = f
}

// SetLog 设置默认日志
func SetLog(log Logger) {
	Log = log
	SetLogLevel(LogLevel)
}

// SetLogLevel 设置默认日志级别
func SetLogLevel(level int) {
	LogLevel = level
	if Log != nil {
		Log.SetLevel(LogLevel)
	}
}

// New 返回可操作数据库
func New(s string, arg string) (db Database, err error) {
	if f, ok := hashes[s]; ok {
		return f(arg)
	}
	return nil, ErrDbParamsInvalid
}

// default
func defaultHookParseSafe(m map[string]interface{}, whiteList map[string]interface{}, blackList map[string]interface{},
	defaultVals map[string]interface{}) (err error) {
	err = ErrHookFuncUndefined
	return
}
func defaultHookParseMgo(m map[string]interface{}) (d map[string]interface{}, err error) {
	err = ErrHookFuncUndefined
	return
}

// init
func init() {
	// utils
	if err := initUtils(); err != nil {
		panic(err)
	}
}
