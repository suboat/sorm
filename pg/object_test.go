package postgres

import (
	"github.com/suboat/sorm"
	"testing"
	"time"
)

func Test_ObjectCreate(t *testing.T) {
	var (
		m   = testGetDbMust().Model("demoperson")
		p   = &demoPerson{}
		err error
	)
	if err = m.EnsureIndexWithTag(p); err != nil {
		t.Fatal(err.Error())
		return
	}

	p.Uid = m.NewUid()
	p.Age = 16
	p.FirstName = "Tomy"
	p.LastName = "Lee"
	p.Message = "talk something"
	p.Weight = 80
	p.Birthday = time.Now()
	p.Address.Number = 5
	p.Address.Road = "nanning"
	p.Password = []byte("somebytes")
	if err = m.Objects().Create(p); err != nil {
		t.Fatal(err.Error())
		return
	}

	// test transaction create
	var tx orm.Trans
	if tx, err = m.Begin(); err != nil {
		t.Fatal(err.Error())
		return
	}
	p.Age += 1
	p.Uid = m.NewUid()
	p.Birthday = time.Now()
	if err = m.Objects().TCreate(p, tx); err != nil {
		t.Error(err.Error())
	}
	// autotrans: auto commit or rollback
	if err = m.AutoTrans(tx); err != nil {
		t.Fatal(err.Error())
		return
	}

	// count
	n := 0
	if n, err = m.Objects().Count(); err != nil {
		t.Fatal(err.Error())
		return
	} else {
		println("count1:", n)
	}
	if n, err = m.Objects().Filter(orm.M{
		"age": 16,
	}).Count(); err != nil {
		t.Error(err.Error())
		return
	} else {
		println("count2:", n)
	}

	// fetch
	var pLis []demoPerson
	if err = m.Objects().Skip(1).All(&pLis); err != nil {
		t.Error(err.Error())
		return
	} else {
		println("fetch1:", len(pLis))
	}

	// filter and meta
	pLis2 := []demoPerson{}
	m2 := m.Objects().Filter(orm.M{
		"age":      16,
		"lastname": "Lee",
	}).Sort("-age").Limit(10)
	if err = m2.All(&pLis2); err != nil {
		t.Error(err.Error())
		return
	} else {
		println("fetch2:", len(pLis2))
	}
	var meta *orm.Meta
	if meta, err = m2.Meta(); err != nil {
		t.Error(err.Error())
		return
	} else {
		println("meta:", meta.Length, meta.Total, meta.Limit, meta.Skip)
	}

	// one
	var pOne demoPerson
	if err = m.Objects().Filter(orm.M{"uid": p.Uid}).One(&pOne); err != nil {
		t.Error(err.Error())
		return
	} else {
		println("one:", pOne.FirstName, pOne.Uid.String(), pOne.Birthday.String(), pOne.Address.Road, string(pOne.Password))
	}

	// update
	if err = m.Objects().Update(map[string]interface{}{
		"height": 170,
	}); err != nil {
		t.Error(err.Error())
		return
	}

	// update one
	pOne.Height = 180
	if err = m.Objects().Filter(orm.M{"uid": p.Uid}).UpdateOne(pOne); err != nil {
		t.Error(err.Error())
		return
	}

	// update trans
	if tx, err = m.Begin(); err != nil {
		t.Fatal(err.Error())
		return
	}
	if err = m.Objects().TUpdate(map[string]interface{}{
		"message": "talk nothing",
	}, tx); err != nil {
		println("update trans error:", err.Error())
	}
	if err = m.AutoTrans(tx); err != nil {
		t.Fatal(err.Error())
		return
	}

	// update trans one
	pOne.Weight = 90
	if tx, err = m.Begin(); err != nil {
		t.Fatal(err.Error())
		return
	}
	if err = m.Objects().Filter(orm.M{"uid": p.Uid}).TUpdate(pOne, tx); err != nil {
		println("update trans one error:", err.Error())
	}
	if err = m.Objects().Filter(orm.M{"uid": p.Uid}).TUpdateOne(map[string]interface{}{
		"address": map[string]interface{}{
			"Road":   "gaoxin",
			"Number": 3,
		},
	}, tx); err != nil {
		println("update[map] trans one error:", err.Error())
		return
	}
	if err = m.AutoTrans(tx); err != nil {
		t.Fatal(err.Error())
		return
	}

	// delete
	if n, err = m.Objects().Count(); err != nil {
		t.Error(err.Error())
		return
	}
	if n >= 6 {
		if err = m.Objects().Filter(orm.M{
			"age": 16,
		}).Delete(); err != nil {
			t.Fatal(err.Error())
			return
		}
		// delete trans
		if tx, err = m.Begin(); err != nil {
			t.Fatal(err.Error())
			return
		}
		if err = m.Objects().Filter(orm.M{"age": 17}).TDelete(tx); err != nil {
			println("delete trans one error:", err.Error())
		}
		if err = m.AutoTrans(tx); err != nil {
			t.Fatal(err.Error())
			return
		}
	}
}
