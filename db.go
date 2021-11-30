package orm

import (
	"time"
)

// ArgModel 取model参数
type ArgModel struct {
	// 事务默认超时
	TransTimeout time.Duration
	// 日志级别
	LogLevel int
	// 自定义表的SQL语句
	Sql string
}

// Database 数据库对象
type Database interface {
	//
	String() string     // 数据库类型及版本
	DriverName() string // 数据库类型
	Close() error       // 可关闭
	//
	Model(table string) Model                    // 获取table或者collection
	ModelWith(table string, opt *ArgModel) Model // 获取table或者collection
	// 设置默认值
}
