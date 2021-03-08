package mssql

import (
	"github.com/stretchr/testify/require"
	"github.com/suboat/sorm"
	"testing"
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
	as := require.New(t)
	db, err := NewDb(testConn)
	as.Nil(err)
	testDB = db
	return testDB
}

type Eukaryota struct {
	TaxonomyID    int    `sorm:"index" json:"taxonomyId"`    // 物种编号 9606
	Chromosomes   int32  `sorm:"index" json:"chromosomes"`   // 染色体数目
	SpeciesNumber uint64 `sorm:"index" json:"speciesNumber"` // 全球物种数 6,000,000,000
	//IndividualCells types.BigInt `sorm:"decimal(16);index" json:"individualCells"` // 个体组成细胞数 30,000,000,000,000
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

// 建表
func Test_Model(t *testing.T) {
	db := testServe(t)
	as := require.New(t)
	//
	if db != nil {
		m := db.Model("test_table").With(&orm.ArgModel{LogLevel: orm.LevelDebug})
		as.Nil(m.Ensure(&Eukaryota{}))
	}
}
