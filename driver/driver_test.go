package driver

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"

	_ "github.com/suboat/sorm/driver/mongo"
	_ "github.com/suboat/sorm/driver/mssql"
	_ "github.com/suboat/sorm/driver/mysql"
	_ "github.com/suboat/sorm/driver/pg"
	_ "github.com/suboat/sorm/driver/sqlite"

	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	// TestName 要测试的数据库
	TestName string
	testDir  string
	testDB   orm.Database
)

// 取测试运行的目录
func testGetDir() string {
	if len(testDir) == 0 {
		_, filename, _, _ := runtime.Caller(0)
		testDir = filepath.Dir(filename)
	}
	return testDir
}

// 读取数据库配置
func testGetDB() orm.Database {
	if len(TestName) == 0 {
		//TestName = orm.DriverNamePostgres
		//TestName = orm.DriverNameMysql
		TestName = orm.DriverNameMsSql
		//TestName = orm.DriverNameSQLite
		//TestName = orm.DriverNameMongo
	}
	if testDB != nil {
		return testDB
	}
	testGetDir()

	// 开启调试日志级别
	//orm.SetLogLevel(orm.LevelDebug)
	//orm.Log = log.NewLogFile(filepath.Join(testDir, "test.log"))
	//orm.SetLogLevel(orm.LogLevel)

	// pg
	var (
		conn string
	)
	switch TestName {
	case orm.DriverNamePostgres:
		conn = `{"user":"business", "password": "business", "host": "127.0.0.1", "port": "65432", "database": "business"}`
	case orm.DriverNameMysql:
		conn = `{"user":"tester", "password": "business", "host": "192.168.6.6", "port": "3306", "database": "tester_sorm"}`
	case orm.DriverNameMsSql:
		conn = `{"user":"tester", "password": "business", "host": "192.168.6.6", "port": "1433", "database": "tester_main"}`
	case orm.DriverNameSQLite:
		conn = `{"database":"data_sqlite/business.db"}`
	case orm.DriverNameMongo:
		conn = `{"url":"mongodb://127.0.0.1:27017/", "db": "business"}`

	default:
		panic(fmt.Errorf(`unknown database "%s"`, TestName))
	}

	if db, err := orm.New(TestName, conn); err == nil {
		testDB = db
	} else {
		panic(err)
	}

	return testDB
}

// TestMain
func TestMain(m *testing.M) {
	// 链接数据库
	log.Infof("databases %v opened.", testGetDB())

	// 运行测试
	m.Run()

	// 关闭数据库
	log.Infof("databases %v closing...", testGetDB())
	if err := testGetDB().Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}
