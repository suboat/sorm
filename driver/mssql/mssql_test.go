package mssql

import (
	"github.com/stretchr/testify/require"
	"github.com/suboat/sorm"
	"testing"
	"time"
)

var (
	testConn = `{"user":"tester", "password": "business", "host": "192.168.6.6", "port": "1433", "database": "tester_main"}`
	testDB   orm.Database
)

//
func testServe(t *testing.T) (ret orm.Database) {
	if testDB != nil {
		return testDB
	}
	CfgDbUnsafe = true
	as := require.New(t)
	db, err := NewDb(testConn)
	as.Nil(err)
	testDB = db
	return testDB
}

//
func Test_Version(t *testing.T) {
	db := testServe(t)
	as := require.New(t)
	if s := db.(*DatabaseSQL); s != nil {
		var version string
		as.Nil(s.DB.Get(&version, `select @@version`))
		t.Logf(`%s`, version)
	}
}

//
func Test_Query(t *testing.T) {
	db := testServe(t)
	as := require.New(t)
	//
	type user struct {
		CardId  string     `db:"card_id"`  //
		VipName string     `db:"vip_name"` //
		VipDate *time.Time `db:"vip_date"` //
	}
	//
	if true {
		var data []*user
		as.Nil(db.Model("t_rm_vip_info").Objects().Sort("card_id").Limit(10).All(&data))
		for i, v := range data {
			t.Logf(`[all1] #%d/%d %s`, i+1, len(data), orm.JSONMust(v))
		}
	}
	if true {
		var data []*user
		as.Nil(db.Model("t_rm_vip_info").ObjectsWith(&orm.ArgObjects{
			LogLevel: orm.LevelDebug,
		}).Sort("card_id").Limit(10).Filter(orm.M{
			"vip_date": nil,
		}).All(&data))
		for i, v := range data {
			t.Logf(`[all2] #%d/%d %s`, i+1, len(data), orm.JSONMust(v))
		}
	}
}
