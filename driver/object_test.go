package driver

import (
	"github.com/suboat/sorm"

	"testing"

	"database/sql/driver"
	"encoding/json"
)

// User 用户
type User struct {
	Username string   `sorm:"primary;size(36)" json:"username"`    // 用户
	Amount   float64  `sorm:"decimal(20,8);index" json:"amount"`   // 总额:余额+冻结
	Balance  float64  `sorm:"decimal(20,8);index" json:"balance"`  // 余额
	Freezing float64  `sorm:"decimal(20,8);index" json:"freezing"` // 冻结
	Meta     UserMeta `sorm:"" json:"meta"`                        //
}

// UserMeta
type UserMeta struct {
	Country string
	City    string
}

//
func (d UserMeta) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *UserMeta) Scan(src interface{}) (err error) {
	return json.Unmarshal(src.([]byte), d)
}

// 新建记录与更新
func Test_ObjectsCreateUpdate(t *testing.T) {
	db := testGetDB()
	var (
		tbl0 = "test_object_person"
		m0   = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelInfo})
		err  error
	)
	if err = m0.Drop(); err != nil {
		t.Fatal(err)
		return
	}
	if err = m0.Ensure(&User{}); err != nil {
		t.Fatal(err)
		return
	}

	// 新建用户
	var (
		user0 = &User{
			Username: "tester0",
			Amount:   1000,
			Meta: UserMeta{
				Country: "China",
				City:    "Nanning",
			},
		}
		user1 = &User{
			Username: "tester1",
			Amount:   500,
			Meta: UserMeta{
				Country: "China",
				City:    "Guilin",
			},
		}
	)
	if err = m0.Objects().Create(user0); err != nil {
		t.Fatal(err)
		return
	}
	if err = m0.Objects().Create(user1); err != nil {
		t.Fatal(err)
		return
	}

	// 更新用户: 用结构体
	user0.Amount = 2000
	user0.Meta.City = "Liuzhou"
	if err = m0.Objects().With(
		&orm.ArgObjects{LogLevel: orm.LevelDebug}).Filter(orm.M{"username": user0.Username}).UpdateOne(
		user0,
	); err == nil {
		_user := new(User)
		if err = m0.Objects().Filter(orm.M{"username": user0.Username}).One(_user); err != nil {
			t.Fatal(err)
			return
		}
		if orm.JSONMust(user0) != orm.JSONMust(_user) {
			t.Fatalf(`"%v" update after diff:
%s
%s`, db, orm.JSONMust(user0), orm.JSONMust(_user))
			return
		}
	} else {
		t.Fatal(err)
		return
	}

	// 更新用户: 用map
	user0.Amount = 3000
	user0.Meta.City = "Guilin"
	if err = m0.Objects().With(
		&orm.ArgObjects{LogLevel: orm.LevelDebug}).Filter(orm.M{"username": user0.Username}).UpdateOne(
		map[string]interface{}{
			"Amount": user0.Amount,
			"Meta":   user0.Meta,
		}); err == nil {
		_user := new(User)
		if err = m0.Objects().Filter(orm.M{"username": user0.Username}).One(_user); err != nil {
			t.Fatal(err)
			return
		}
		if orm.JSONMust(user0) != orm.JSONMust(_user) {
			t.Fatalf(`"%v" update after diff:
%s
%s`, db, orm.JSONMust(user0), orm.JSONMust(_user))
			return
		}
	} else {
		t.Fatal(err)
		return
	}

	// 确保user1没有被改动
	_user1 := new(User)
	if err = m0.Objects().Filter(orm.M{"username": user1.Username}).One(_user1); err != nil {
		t.Fatal(err)
		return
	}
	if orm.JSONMust(user1) != orm.JSONMust(_user1) {
		t.Fatalf(`"%v" update after diff:
%s
%s`, db, orm.JSONMust(user1), orm.JSONMust(_user1))
		return
	}

	// log
	t.Logf(`"%v" PASS %s`, db, orm.JSONMust(user0))
}

// 压测更新
func Benchmark_ObjectsUpdate(b *testing.B) {
	db := testGetDB()
	b.Logf(`ObjectsUpdate: %v`, db)

	var (
		tbl0 = "bench_object_update"
		m0   = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelError})
		n    = b.N
		err  error
	)
	if err = m0.Drop(); err != nil {
		b.Fatal(err)
		return
	}
	if err = m0.Ensure(&User{}); err != nil {
		b.Fatal(err)
		return
	}
	if err = m0.Objects().Create(&User{
		Username: "bench",
		Amount:   float64(n),
	}); err != nil {
		b.Fatal(err)
		return
	}

	// 数据库加减
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err = m0.Objects().Filter(orm.M{"username": "bench"}).Update(map[string]interface{}{
				orm.TagUpdateInc: map[string]interface{}{
					"amount":   -1,
					"freezing": 1,
				},
			}); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.StopTimer()

	// 验证
	_user := new(User)
	if err = m0.Objects().Filter(orm.M{"username": "bench"}).One(_user); err != nil {
		b.Fatal(err)
		return
	}
	if !(_user.Amount == 0 && _user.Freezing == float64(n)) {
		b.Fatalf(`unexpect result: %f %f`, _user.Amount, _user.Freezing)
		return
	}
}
