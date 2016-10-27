package mongo

import (
	"encoding/json"
	"fmt"
	"github.com/suboat/sorm"
	"gopkg.in/mgo.v2"
)

type db struct {
	session *mgo.Session  // mongo session
	db      *mgo.Database // 数据库
	arg                   // 传入参数
}

type arg struct {
	url    string
	dbName string
}

// 初始或重设数据库连接
func (d *db) reset() (err error) {
	if d.session != nil {
		d.Close()
	}
	d.session, err = mgo.Dial(d.url)
	if err == nil {
		// Optional. Switch the session to a monotonic behavior.
		d.session.SetMode(mgo.Monotonic, true)
	}
	d.db = d.session.DB(d.dbName)
	return
}

// Model
func (d *db) Model(s string) orm.Model {
	m := new(model)
	m.name = s
	m.collection = d.db.C(s)
	return m
}

// 关闭数据库连接
func (d *db) Close() (err error) {
	if d.session != nil {
		d.session.Close()
	}
	return
}

// 可打印
func (d *db) String() string {
	return fmt.Sprintf("mongo:%s->%s", d.url, d.dbName)
}

// 解析数据库链接参数
func parserArg(s string) (a *db, err error) {
	if len(s) == 0 {
		err = orm.ErrParamsType
		return
	}
	a = new(db)
	// 解析json参数
	arg := map[string]string{}
	json.Unmarshal([]byte(s), &arg)
	// 兼容sql的参数读取
	a.url = arg["url"]
	if len(a.url) == 0 {
		a.url = arg["host"]
	}
	a.dbName = arg["db"]
	if len(a.dbName) == 0 {
		a.dbName = arg["dbname"]
	}
	return
}

func NewDb(arg string) (orm.Database, error) {
	d, err := parserArg(arg)
	// 数据库初始化
	if err == nil {
		err = d.reset()
	}
	return d, err
}

func init() {
	orm.RegisterHash(orm.Mongo, NewDb)
}
