package postgres

import (
	"git.yichui.net/open/orm"
	"testing"
)

var (
	testConfig = `{"user":"suboat", "password": "suboat123", "host": "192.168.199.188", "dbname": "suboat", "sslmode": "disable"}`
	testDb     orm.Database
)

func testGetDbMust() orm.Database {
	if testDb != nil {
		return testDb
	}
	var (
		err error
	)
	if testDb, err = NewDb(testConfig); err != nil {
		panic(err)
	}
	return testDb
}

func Test_NewDb(t *testing.T) {
	var (
		err error
	)
	if testDb, err = NewDb(testConfig); err != nil {
		t.Fatal(err.Error())
	} else {
		println(testDb.String())
	}
}
