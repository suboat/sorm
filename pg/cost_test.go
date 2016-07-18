package postgres

import (
	"fmt"
	"github.com/suboat/sorm"
	"testing"
	"time"
)

var (
	modelWallet     = testGetDbMust().Model("wallet")
	modelWalletflow = testGetDbMust().Model("walletflow")
	txTimeOut       = time.Second * 10 // 事务操作超时
	err             error
)

func testWalletEnsure(username string) (err error) {
	var c int
	if c, err = modelWallet.Objects().Filter(map[string]interface{}{
		"username": username,
	}).Count(); err != nil {
		return
	} else if c == 0 {
		if err = modelWallet.Objects().Create(&wallet{
			Username: username,
			Balance:  0,
			Amount:   0,
		}); err != nil {
			return
		}
		if err = testWalletflowUpdateWallet(&walletFlow{
			Username: username,
			IsIncome: true,
			Amount:   1000,
		}, nil); err != nil {
			return
		}
	}
	return
}

func testWalletflowUpdateWallet(wf *walletFlow, tx orm.Trans) (err error) {
	var (
		w *wallet = new(wallet)
	)

	if tx == nil {
		if tx, err = modelWallet.Begin(); err != nil {
			println("modelWallet.Begin()", err.Error())
			return
		}
		defer modelWallet.AutoTrans(tx)
	} else {
		//// 超时回滚
		//t := time.NewTimer(txTimeOut)
		//go func() {
		//	defer func() {
		//		if _err := recover(); _err != nil {
		//			fmt.Printf("panic %v\n", _err)
		//		}
		//	}()
		//
		//	<-t.C
		//	// timeout
		//	println("timeout")
		//	if tx == nil {
		//		fmt.Printf("tx == nil\n")
		//	} else {
		//		if er := tx.Rollback(); er != nil {
		//			fmt.Printf("Rollback %v\n", er)
		//		}
		//	}
		//}()
	}

	// 行锁
	//if _, err = tx.Exec("SELECT FROM wallet WHERE username = $1 FOR UPDATE", wf.Username); err != nil {
	//	//println("lock error", err.Error())
	//	return
	//}
	//if err = modelWallet.Objects().Filter(orm.M{"username": wf.Username}).TLockUpdate(tx); err != nil {
	//	println("lock error", err.Error())
	//	return
	//}

	// 读钱包
	//var tLis = []*wallet{}
	//if err = modelWallet.Objects().Filter(orm.M{"username": wf.Username}).TAll(&tLis, tx); err != nil {
	//	//if err = modelWallet.Objects().Filter(orm.M{"username": wf.Username}).All(&tLis); err != nil {
	//	println("fetch err", err.Error())
	//	return
	//} else {
	//	w = tLis[0]
	//}
	if err = modelWallet.Objects().Filter(orm.M{"username": wf.Username}).TOne(w, tx); err != nil {
		return
	}

	// 创建流水
	if wf.IsIncome == true {
		w.Amount += wf.Amount
		w.Balance += wf.Amount
	} else {
		w.Amount -= wf.Amount
		w.Balance -= wf.Amount
	}
	wf.Balance = w.Balance
	if err = modelWalletflow.Objects().TCreate(wf, tx); err != nil {
		return
	}

	// 更新
	if wf.IsIncome == false {
		if w.Balance < 0 {
			err = fmt.Errorf("balance not enough")
			tx.ErrorSet(err)
			return
		}
		//if _res, _err := tx.Exec("UPDATE wallet SET balance = balance - $1 WHERE username = $2 AND balance > 0",
		//if _, err = tx.Exec("UPDATE wallet SET balance = balance - $1 WHERE username = $2",
		//	wf.Amount, wf.Username); err != nil {
		//	println("UPDATE err", err.Error())
		//	return
		//}
		//
		if err = modelWallet.Objects().Filter(orm.M{"username": wf.Username}).TUpdate(map[string]interface{}{
			orm.TagUpdateInc: map[string]interface{}{
				"balance": -wf.Amount,
				"amount":  -wf.Amount,
			},
		}, tx); err != nil {
			return
		}
	} else {
		// 打款
		//if _, err = tx.Exec("UPDATE wallet SET balance = balance + $1 WHERE username = $2",
		//	wf.Amount, wf.Username); err != nil {
		//	println("UPDATE2 err", err.Error())
		//	return
		//}
		if err = modelWallet.Objects().Filter(orm.M{"username": wf.Username}).TUpdate(map[string]interface{}{
			orm.TagUpdateInc: map[string]interface{}{
				"balance": wf.Amount,
				"amount":  wf.Amount,
			},
		}, tx); err != nil {
			return
		}
	}

	return
}

// 测试并发消费
func Test_CostConcurrency(t *testing.T) {
	// init
	if err = modelWallet.EnsureIndexWithTag(&wallet{}); err != nil {
		t.Fatal(err)
	}
	if err = modelWalletflow.EnsureIndexWithTag(&walletFlow{}); err != nil {
		t.Fatal(err)
	}
	if err = testWalletEnsure("pepole1"); err != nil {
		t.Fatal(err)
		return
	}
	if err = testWalletEnsure("pepole2"); err != nil {
		t.Fatal(err)
		return
	}
	if err = testWalletEnsure("pepole3"); err != nil {
		t.Fatal(err)
		return
	}

	// 拉取流水
	if false {
		var wfLis = []*walletFlow{}
		if err = modelWalletflow.Objects().All(&wfLis); err != nil {
			t.Fatal(err)
		}
		for _, d := range wfLis {
			t.Log(d.Id, d.Balance)
		}
		return
	}

	// 并发测试
	cNum := 2
	sem := make(chan bool, cNum)
	for i := 0; i < cNum; i++ {
		sem <- true
	}
	start := time.Now()
	for i := 0; i < 1100; i++ {
		go func(i2 int) {
			<-sem
			defer func() { sem <- true }()
			tx, _ := modelWallet.Begin()
			defer modelWallet.AutoTrans(tx)
			// lock
			if err = modelWallet.Objects().Filter(orm.M{
				orm.TagQueryKeyOr: []interface{}{
					map[string]interface{}{"username": "pepole1"},
					map[string]interface{}{"username": "pepole2"},
				},
			}).TLockUpdate(tx); err != nil {
				//println("lock error", err.Error())
				return
			}
			var wf = new(walletFlow)
			wf.Username = "pepole1"
			wf.IsIncome = false
			wf.Amount = 1
			// 扣款
			if err = testWalletflowUpdateWallet(wf, tx); err != nil {
				println(i2, err.Error())
				return
			} else {
				println(i2, "A ok")
			}
			// 打款
			wf.Username = "pepole2"
			wf.IsIncome = true
			if err = testWalletflowUpdateWallet(wf, tx); err != nil {
				println(i2, "A", err.Error())
				return
			}

		}(i)

		go func(i2 int) {
			<-sem
			defer func() { sem <- true }()
			tx, _ := modelWallet.Begin()
			defer modelWallet.AutoTrans(tx)
			// lock
			if err = modelWallet.Objects().Filter(orm.M{
				orm.TagQueryKeyOr: []interface{}{
					map[string]interface{}{"username": "pepole1"},
					map[string]interface{}{"username": "pepole2"},
				},
			}).TLockUpdate(tx); err != nil {
				//println("lock error", err.Error())
				return
			}
			var wf = new(walletFlow)
			wf.Username = "pepole2"
			wf.IsIncome = false
			wf.Amount = 1
			// 扣款
			if err = testWalletflowUpdateWallet(wf, tx); err != nil {
				println(i2, err.Error())
				return
			} else {
				println(i2, "B ok")
			}
			// 打款
			wf.Username = "pepole1"
			wf.IsIncome = true
			if err = testWalletflowUpdateWallet(wf, tx); err != nil {
				println(i2, "B", err.Error())
				return
			}
		}(i)

		go func(i2 int) {
			<-sem
			defer func() { sem <- true }()
			tx, _ := modelWallet.Begin()
			defer modelWallet.AutoTrans(tx)
			// lock
			if err = modelWallet.Objects().Filter(orm.M{
				orm.TagQueryKeyOr: []interface{}{
					map[string]interface{}{"username": "pepole1"},
					map[string]interface{}{"username": "pepole3"},
				},
			}).TLockUpdate(tx); err != nil {
				//println("lock error", err.Error())
				return
			}
			var wf = new(walletFlow)
			wf.Username = "pepole3"
			wf.IsIncome = false
			wf.Amount = 1
			// 扣款
			if err = testWalletflowUpdateWallet(wf, tx); err != nil {
				println(i2, err.Error())
				return
			} else {
				println(i2, "C ok")
			}
			// 打款
			wf.Username = "pepole1"
			wf.IsIncome = true
			if err = testWalletflowUpdateWallet(wf, tx); err != nil {
				println(i2, "B", err.Error())
				return
			}
		}(i)
	}
	println("time: ", time.Now().Sub(start).Seconds())

	time.Sleep(time.Second * 1000)
}
