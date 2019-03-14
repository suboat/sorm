package driver

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/driver/mongo"

	"testing"

	"database/sql/driver"
	"encoding/json"
	"math/rand"
	"time"
)

// 事务测试

// WalletFlow
type WalletFlow struct {
	ID         uint64         `sorm:"serial;primary" json:"id" bson:"-" ` // id
	Username   string         `sorm:"size(36);index" json:"username"`     // 用户
	IsIncome   bool           `sorm:"index" json:"isIncome"`              // true: 收入流水
	Amount     float64        `sorm:"decimal(20,8);index" json:"amount"`  // 流水金额
	Time       time.Time      `sorm:"index" json:"time"`                  // 时间
	CommentTag string         `sorm:"size(36);index" json:"commentTag"`   // 备注
	Meta       WalletFlowMeta `sorm:"" json:"meta"`                       // 拓展信息
}

// WalletFlowMeta
type WalletFlowMeta struct {
	Comment          string  `json:"comment"`          // 备注
	SnapshotBalance  float64 `json:"snapshotBalance"`  // 余额快照
	SnapshotFreezing float64 `json:"snapshotFreezing"` // 冻结快照
}

func (d WalletFlowMeta) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *WalletFlowMeta) Scan(src interface{}) (err error) {
	return json.Unmarshal(src.([]byte), d)
}

// 压测事务 FIXME: mongoDB不运行此测试
func Benchmark_Trans(b *testing.B) {
	db := testGetDB()
	b.Logf(`Benchmark_Trans: %v`, db)

	var (
		tbl0  = "bench_trans_person"
		tbl1  = "bench_trans_flow"
		m0    = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelFatal})
		m1    = db.Model(tbl1).With(&orm.ArgModel{LogLevel: orm.LevelFatal})
		err   error
		limit chan bool
	)
	if err = m0.Drop(); err != nil {
		b.Fatal(err)
		return
	}
	if err = m0.Ensure(&User{}); err != nil {
		b.Fatal(err)
		return
	}
	if err = m1.Drop(); err != nil {
		b.Fatal(err)
		return
	}
	if err = m1.Ensure(&WalletFlow{}); err != nil {
		b.Fatal(err)
		return
	}

	// 新建用户
	var (
		rn     = rand.New(rand.NewSource(0))
		unit   = 1.0
		amount = float64(b.N)
		users  = []*User{
			{
				Username: "trans0",
				Amount:   amount,
				Balance:  amount,
				Meta: UserMeta{
					Country: "China",
					City:    "Nanning",
				},
			},
			{
				Username: "trans1",
				Amount:   amount,
				Balance:  amount,
				Meta: UserMeta{
					Country: "China",
					City:    "Liuzhou",
				},
			},
			{
				Username: "trans2",
				Amount:   amount,
				Balance:  amount,
				Meta: UserMeta{
					Country: "China",
					City:    "Guilin",
				},
			},
		}
	)
	for _, u := range users {
		if err = m0.Objects().Create(u); err != nil {
			b.Fatal(err)
			return
		}
	}

	// mongo
	if db.DriverName() == orm.DriverNameMongo {
		mongo.CfgTxUnsafe = true
		limit = make(chan bool, 1) // 限制并发数
	}

	// debug 限制并发数
	//limit = make(chan bool, 2)

	// 并发测试
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 限制并发数
			if limit != nil {
				limit <- true
			}
			var (
				uFrom  = &User{}
				uTo    = &User{}
				wfFrom = &WalletFlow{
					Amount:     unit,
					IsIncome:   false,
					CommentTag: "bench_trans",
				}
				wfTo = &WalletFlow{
					Amount:     unit,
					IsIncome:   true,
					CommentTag: "bench_trans",
				}
				now time.Time
				tx  orm.Trans
				err error
				n   = rn.Int()
			)

			switch n % 6 {
			case 0:
				// user0给user1转钱
				uFrom.Username = "trans0"
				uTo.Username = "trans1"
			case 1:
				// user0给user2转钱
				uFrom.Username = "trans0"
				uTo.Username = "trans2"
			case 2:
				// user1给user0转钱
				uFrom.Username = "trans1"
				uTo.Username = "trans0"
			case 3:
				// user1给user2转钱
				uFrom.Username = "trans1"
				uTo.Username = "trans2"
			case 4:
				// user2给user0转钱
				uFrom.Username = "trans2"
				uTo.Username = "trans0"
			case 5:
				// user2给user1转钱
				uFrom.Username = "trans2"
				uTo.Username = "trans1"
			default:
				b.Fatalf("unknown error %d", n)
			}
			wfFrom.Username = uFrom.Username
			wfTo.Username = uTo.Username

			// 事务开始
			// if tx, err = m0.Begin(); err != nil {
			if tx, err = m0.BeginWith(&orm.ArgTrans{Level: orm.TransRepeatableRead}); err != nil {
				b.Fatal(err)
				break
			}

			// 行锁
			if err = m0.Objects().Filter(orm.M{
				orm.TagQueryKeyOr: []interface{}{
					map[string]interface{}{"username": uFrom.Username},
					map[string]interface{}{"username": uTo.Username},
				},
			}).TLockUpdate(tx); err != nil {
				tx.ErrorSet(err)
				goto FINAL
			}

			// 手机快照信息
			if err = m0.Objects().Filter(orm.M{"username": uFrom.Username}).TOne(uFrom, tx); err != nil {
				tx.ErrorSet(err)
				goto FINAL
			}
			if err = m0.Objects().Filter(orm.M{"username": uTo.Username}).TOne(uTo, tx); err != nil {
				tx.ErrorSet(err)
				goto FINAL
			}
			wfFrom.Meta.SnapshotBalance = orm.FloatAdd(uFrom.Amount, -unit)
			wfTo.Meta.SnapshotBalance = orm.FloatAdd(uTo.Amount, unit)

			// 创建快照
			now = time.Now()
			wfFrom.Time = now
			wfTo.Time = now
			if err = m1.Objects().TCreate(wfFrom, tx); err != nil {
				tx.ErrorSet(err)
				goto FINAL
			}
			if err = m1.Objects().TCreate(wfTo, tx); err != nil {
				tx.ErrorSet(err)
				goto FINAL
			}

			// 扣钱
			if err = m0.Objects().Filter(orm.M{"username": uFrom.Username}).TUpdateOne(map[string]interface{}{
				orm.TagUpdateInc: map[string]interface{}{
					"amount":  -unit,
					"balance": -unit,
				},
			}, tx); err != nil {
				tx.ErrorSet(err)
				goto FINAL
			}

			// 打钱
			if err = m0.Objects().Filter(orm.M{"username": uTo.Username}).TUpdateOne(map[string]interface{}{
				orm.TagUpdateInc: map[string]interface{}{
					"amount":  unit,
					"balance": unit,
				},
			}, tx); err != nil {
				tx.ErrorSet(err)
				goto FINAL
			}

		FINAL:
			// 事务结束
			//if tx.Error() != nil {
			//	//log.Error(tx.Error())
			//	println("tx err:", tx.Error().Error())
			//}
			if err = m0.AutoTrans(tx); err != nil {
				b.Fatal(err)
				break
			}

			// 限制并发数
			if limit != nil {
				<-limit
			}
		}
	})
	b.StopTimer()

	// 验证结果
	var (
		numAll int
	)
	for _, u := range users {
		if err = m0.Objects().Filter(orm.M{"username": u.Username}).One(u); err != nil {
			return
		}
		if _n, _err := m1.Objects().Filter(orm.M{"username": u.Username, "isIncome": false}).Count(); _err == nil {
			numAll += _n
		} else {
			b.Fatal(_err)
			return
		}

		var wfLis []*WalletFlow
		if err = m1.Objects().Filter(orm.M{"username": u.Username}).Sort("id").All(&wfLis); err != nil {
			b.Fatal(err)
			return
		}
		if len(wfLis) > 0 {
			if _last := wfLis[len(wfLis)-1]; _last.Meta.SnapshotBalance != u.Balance {
				b.Log(orm.JSONMust(u))
				b.Log(orm.JSONMust(_last))
				b.Fatalf(`unexspect result %s %f %f`, u.Username, u.Balance, _last.Meta.SnapshotBalance)
			}
		}

		// 验证顺序正确性
		for _i, _wf := range wfLis {
			if _i == 0 {
				continue
			}
			amount := _wf.Amount
			if _wf.IsIncome {
				amount = -amount
			}
			balanceBefore := orm.FloatAdd(_wf.Meta.SnapshotBalance, amount)
			balanceNow := wfLis[_i-1].Meta.SnapshotBalance
			if balanceBefore != balanceNow {
				b.Fatalf(`unexspect result %s %d %f %f`, u.Username, _wf.ID, balanceBefore, balanceNow)
				return
			}
		}
	}
	if _numAll, _err := m1.Objects().Filter(orm.M{"isIncome": false}).Count(); _err != nil {
		b.Fatal(_err)
		return
	} else if _numAll != numAll {
		b.Fatalf(`unexspect result %d %d`, _numAll, numAll)
		return
	}

	// log
	b.Logf(`SUCCESS %d/%d`, numAll, b.N)
}
