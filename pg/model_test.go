package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"git.yichui.net/open/orm"
	"testing"
	"time"
)

type demoAnimal struct {
	Height float32 `sorm:"index"`
	Weight float32
}

type demoPerson struct {
	Uid        orm.Uid          `sorm:"size(36) unique"`
	demoAnimal `bson:",inline"` // mgo default embed fields

	FirstName string    `sorm:"size(32) index"`
	LastName  string    `sorm:"size(64) index"`
	Age       int       `sorm:"index"`
	Birthday  time.Time `sorm:"index"`
	Message   string
	Address   demoAddress
	Password  []byte
}

type demoAddress struct {
	Road   string
	Number int
}

func (m demoAddress) Value() (driver.Value, error) {
	return json.Marshal(m)
	//var (
	//	d   []byte
	//	err error
	//)
	//if d, err = json.Marshal(m); err != nil {
	//	return d, err
	//}
	//return d, err
}

func (m *demoAddress) Scan(src interface{}) (err error) {
	return json.Unmarshal(src.([]byte), m)
	//if err = json.Unmarshal(src.([]byte), m); err != nil {
	//	return
	//}
	//return
}

func Test_ModelEnsureIndexWithTag(t *testing.T) {
	var (
		m   = testGetDbMust().Model("demoperson")
		p   = &demoPerson{}
		err error
	)
	if err = m.EnsureIndexWithTag(p); err != nil {
		t.Fatal(err.Error())
	} else {
		println(m.String())
	}
}
