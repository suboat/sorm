package mongo

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/suboat/sorm"
	"testing"
	"time"
)

type DemoAnimal struct {
	Height float32 `sorm:"index"`
	Weight float32
	Length int64
}

type demoPerson struct {
	Uid        orm.Uid          `sorm:"size(36) unique"`
	DemoAnimal `bson:",inline"` // mgo default embed fields

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

// wallet
type wallet struct {
	Username string  `sorm:"primary size(32)" json:"username"` // 用户
	Amount   float64 `sorm:"index" json:"amount"`              // 总额:余额+冻结
	Balance  float64 `sorm:"index" json:"balance"`             // 余额
	Freezing float64 `sorm:"index" json:"freezing"`            // 冻结
}

// walletFlow
type walletFlow struct {
	Id       int64   `sorm:"serial" json:"id"`               // id
	Username string  `sorm:"index size(32)" json:"username"` // 用户
	IsIncome bool    `sorm:"index" json:"isIncome"`          // true: 收入流水
	Amount   float64 `sorm:"index" json:"amount"`            // 流水金额
	Balance  float64 `sorm:"index" json:"balance"`           // 余额
	//Country    string    `sorm:"index size(16)" json:"country"`  // test
	//CountryNum float64   `sorm:"index" json:"countryNum"`        // test
	//Birthday   time.Time `sorm:"index"`                          // test
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

	if err = testGetDbMust().Model("wallet").EnsureIndexWithTag(&wallet{}); err != nil {
		t.Fatal(err)
	}
	if err = testGetDbMust().Model("walletflow").EnsureIndexWithTag(&walletFlow{}); err != nil {
		t.Fatal(err)
	}
}
