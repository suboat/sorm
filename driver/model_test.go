package driver

import (
	"github.com/stretchr/testify/require"
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/types"

	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// Eukaryota 真核生物
type Eukaryota struct {
	ID              int64        `sorm:"primary;serial" bson:"-" json:"id"`        // 递增ID
	TaxonomyID      int          `sorm:"index" json:"taxonomyId"`                  // 物种编号 9606
	Chromosomes     int32        `sorm:"index" json:"chromosomes"`                 // 染色体数目
	SpeciesNumber   uint64       `sorm:"index" json:"speciesNumber"`               // 全球物种数 6,000,000,000
	IndividualCells types.BigInt `sorm:"decimal(16);index" json:"individualCells"` // 个体组成细胞数 30,000,000,000,000
}

// Vertebrata 脊椎动物
type Vertebrata struct {
	Eukaryota `bson:",inline"` // mgo default embed fields mgo不支持指针类型的继承
	Age       uint32           `sorm:"index" json:"age"`    // 年龄
	Length    float32          `sorm:"index" json:"length"` // 身长
	Weight    float64          `sorm:"index" json:"weight"` // 体重
}

// Mammalia 哺乳动物
type Mammalia struct {
	Vertebrata  `bson:",inline"`
	IsPlacental bool `sorm:"index" json:"isPlacental"` // true: 胎生
}

// Primates 灵长类动物
type Primates struct {
	Mammalia `bson:",inline"`
	Height   float64 `sorm:"decimal(8,4);index" json:"height"` // 身高
}

// Homo 智人
type Homo struct {
	UID      string `sorm:"size(36);unique" json:"uid"` // 唯一ID
	Key      string `sorm:"size(36);unique" json:"key"` // 唯一Key
	Primates `bson:",inline"`
	Birthday time.Time `sorm:"index" json:"birthday"` // 生日
}

// Programmer 程序员
type Programmer struct {
	Homo      `bson:",inline"`
	FirstName string         `sorm:"size(32);unique(lastName)" json:"firstName"` // 名
	LastName  string         `sorm:"size(64);index" lastName:"lastName"`         // 姓
	Echo      string         `sorm:"" json:"echo" db:"alt_name"`                 // 文本类型
	Password  []byte         `sorm:"index" json:"password"`                      // 二进制类型
	MetaJSON  string         `sorm:"json;index" json:"metaJson"`                 // 直接声明json类型
	Order     types.SliceInf `sorm:"index" json:"order"`                         // 数组类型
	Like      types.JSONMap  `sorm:"index" json:"like"`                          // map类型
	Desc      types.JSONText `sorm:"index" json:"desc"`                          // text类型
	Meta      ProgrammerMeta `sorm:"" json:"meta"`                               // 结构体解析为json类型
	// 指针类型,允许null
	Bugs      *string    `sorm:"size(16);index" json:"bugs"` //
	ApplyTime *time.Time `sorm:"index"`                      //
}

// programmerMeta 拓展信息
type ProgrammerMeta struct {
	Habits  []string           `json:"habits"`
	Skills  map[string]float64 `json:"skills"`
	Company string             `json:"company"`
}

func (d ProgrammerMeta) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *ProgrammerMeta) Scan(src interface{}) (err error) {
	return json.Unmarshal(src.([]byte), d)
}

// 建表
func Test_ModelEnsure(t *testing.T) {
	as := require.New(t)
	var (
		db   = testGetDB()
		tbl0 = "test_Eukaryota"
		tbl1 = "test_Programmer"
		m0   = db.Model(tbl0)
		m1   = db.Model(tbl1).With(&orm.ArgModel{LogLevel: orm.LevelInfo})
		err  error
	)
	if err = m0.Drop(); err != nil {
		t.Fatal(err)
	}
	if err = m1.Drop(); err != nil {
		t.Fatal(err)
	}

	// m0 表进化
	if err = m0.Ensure(&Eukaryota{}); err != nil {
		t.Fatal(err)
	}
	if err = m0.Ensure(&Vertebrata{}); err != nil {
		t.Fatal(err)
	}

	if err = m0.Ensure(&Mammalia{}); err != nil {
		t.Fatal(err)
	}
	if err = m0.Ensure(&Primates{}); err != nil {
		t.Fatal(err)
	}
	if err = m0.Ensure(&Homo{}); err != nil {
		t.Fatal(err)
	}
	if err = m0.Ensure(&Programmer{}); err != nil {
		t.Fatal(err)
	}

	// m1 一步到位
	if err = m1.Ensure(&Programmer{}); err != nil {
		t.Fatal(err)
	}
	// 创建记录
	w0 := new(Programmer)
	w0.UID = types.NewUID()
	w0.Key = types.NewAccession()
	w0.TaxonomyID = 9606
	w0.Chromosomes = 46
	w0.SpeciesNumber = 6000000000
	w0.IndividualCells = types.BigInt{Int: *big.NewInt(30000000000000)}
	w0.Age = 18
	w0.Length = 180.9999 // 该值在pg数据库中异常
	w0.Weight = 66.6666
	w0.IsPlacental = true
	w0.Height = 180.9999
	w0.Birthday = time.Now().Truncate(time.Second)
	w0.FirstName = "go"
	w0.LastName = "lang"
	w0.Echo = "hello world"
	w0.Password = []byte("secretBase64")
	w0.MetaJSON = `{"meta": "content"}`
	w0.Order = []interface{}{"develop", "publish"}
	w0.Like = map[string]interface{}{"games": []string{"nintendo", "sony"}}
	w0.Desc = types.JSONText(`{"say": "hello"}`)
	w0.Meta.Habits = []string{"sleep", "again"}
	w0.Meta.Skills = map[string]float64{"golang": 95, "php": 5}
	w0.Meta.Company = "suboat"
	w0.Bugs = &w0.Echo
	w0.ApplyTime = &time.Time{}

	as.Nil(m0.Objects().Create(w0))
	as.Nil(m1.Objects().With(&orm.ArgObjects{LogLevel: orm.LevelDebug}).Create(w0))
	w1 := new(Programmer)
	*w1 = *w0
	w1.UID = types.NewUID()
	w1.Key = types.NewAccession()
	w1.Birthday = time.Now().Truncate(time.Second)

	w1.FirstName = w1.FirstName + "-copy"
	if err = m0.Objects().Create(w1); err != nil {
		t.Fatal(err)
		return
	} else {
		if err = m0.With(&orm.ArgModel{LogLevel: orm.LevelFatal}).Objects().Create(w1); err == nil {
			// firstName-lastName已联合唯一, 此处应该报错
			t.Fatalf(`firstName-lastName unique index fail`)
			return
		}
	}

	as.Nil(m1.Objects().Create(w1))

	// 读记录
	r0 := new(Programmer)
	r1 := new(Programmer)
	r2 := new(Programmer)
	r3 := new(Programmer)
	as.Nil(m1.Objects().Filter(orm.M{"key": w0.Key}).One(r1))
	as.Nil(m0.Objects().With(
		&orm.ArgObjects{LogLevel: orm.LevelDebug}).Filter(orm.M{"key": w0.Key}).One(r0))
	as.Nil(m1.Objects().Filter(orm.M{"key": w1.Key}).One(r3))
	as.Nil(m0.Objects().Filter(orm.M{"key": w1.Key}).One(r2))
	// 检测自增序列
	as.Equal(r0.ID, int64(1))
	as.Equal(r1.ID, int64(1))
	as.Equal(r2.ID, int64(2))
	as.Equal(r3.ID, int64(2))

	// 比较
	if w0.Birthday.Equal(r0.Birthday) {
		w0.Birthday = r0.Birthday // 忽略时区表达差异
	}
	if w1.Birthday.Equal(r2.Birthday) {
		w1.Birthday = r2.Birthday // 忽略时区表达差异
	}
	if w0.ApplyTime.Equal(*r0.ApplyTime) {
		w0.ApplyTime = r0.ApplyTime // 忽略时区表达差异
	}
	if w1.ApplyTime.Equal(*r2.ApplyTime) {
		w1.ApplyTime = r2.ApplyTime // 忽略时区表达差异
	}
	//
	if r0.ID == 1 && w0.ID == 0 {
		// 递增就正确
		w0.ID = r0.ID
	}
	if r2.ID == 2 && r3.ID == 2 && w1.ID == 0 {
		// 递增并且跳过2就正确
		r2.ID = 0
		r3.ID = 0
	}
	if orm.JSONMust(w0) != orm.JSONMust(r0) {
		t.Fatalf(`"%v" write and read diff: 
%s
%s`, db, orm.JSONMust(w0), orm.JSONMust(r0))
		return
	}
	if orm.JSONMust(w0) != orm.JSONMust(r1) {
		t.Fatalf(`"%v" write and read diff: 
%s
%s`, db, orm.JSONMust(w0), orm.JSONMust(r1))
		return
	}
	if orm.JSONMust(w1) != orm.JSONMust(r2) {
		t.Fatalf(`"%v" write and read diff: 
%s
%s`, db, orm.JSONMust(w1), orm.JSONMust(r2))
		return
	}
	if orm.JSONMust(w1) != orm.JSONMust(r3) {
		t.Fatalf(`"%v" write and read diff: 
%s
%s`, db, orm.JSONMust(w1), orm.JSONMust(r3))
		return
	}

	// log
	t.Logf(`"%v" PASS %s`, db, orm.JSONMust(r0))

	// virtual
	switch TestName {
	case orm.DriverNameMongo:
		// 目前无法通过测试
		break
	default:
		testModelVirtual(t)
	}

	// group
	switch TestName {
	case orm.DriverNameMongo:
		// 目前无法通过测试
		break
	default:
		testModelGroup(t)
	}
}

func testModelVirtual(t *testing.T) {
	as := require.New(t)
	var (
		err  error
		db   = testGetDB()
		tbl0 = "test_Eukaryota"
		tbl1 = "test_Relation"
		tbs1 = struct {
			ID  int    `sorm:"index" json:"id"`
			Msg string `sorm:"size(255);index" json:"msg"`
		}{
			ID:  1,
			Msg: "message from relation",
		}
		m0 = db.Model(tbl1)
		//
		m1 = db.ModelWith(tbl0, &orm.ArgModel{
			Sql: fmt.Sprintf(
				`select tb1.id,taxonomyId,chromosomes,speciesNumber,individualCells,msg from %s tb1 LEFT JOIN %s tb2 ON tb2.id = tb1.id`,
				strings.ToLower(tbl0),
				strings.ToLower(tbl1)),
			LogLevel: orm.LevelDebug,
		})
	)
	if err = m0.Drop(); err != nil {
		t.Fatal(err)
	}
	if err = m0.Ensure(tbs1); err != nil {
		t.Fatal(err)
	} else if err = m0.Objects().Create(tbs1); err != nil {
		t.Fatal(err)
	}
	//
	if true {
		var d1 = new(Eukaryota)
		as.Nil(m1.Objects().Filter(orm.M{"id": tbs1.ID}).One(d1))
		t.Logf(`[virtual-table] one %d`, d1.SpeciesNumber)
	}
	if true {
		type JoinObj struct {
			Eukaryota
			Msg *string
		}
		var (
			dl1 []*Eukaryota
			dl2 []*JoinObj
		)
		as.Nil(m1.Objects().Limit(10).All(&dl1))
		t.Logf(`[virtual-table] all1 %d`, len(dl1))
		as.Nil(m1.Objects().Limit(10).All(&dl2))
		t.Logf(`[virtual-table] all2 %d`, len(dl2))

		as.Equal(2, len(dl2))
		as.NotNil(dl2[0].Msg)
		as.Equal(tbs1.Msg, *dl2[0].Msg)
		for _, v := range dl2 {
			if v.Msg != nil {
				t.Logf(`[virtual-table] match #%d %s`, v.ID, *v.Msg)
			}
		}
	}
}

//
func testModelGroup(t *testing.T) {
	as := require.New(t)
	as.Nil(nil)
	var (
		db   = testGetDB()
		tbl0 = "test_Programmer"
		gps  = []string{
			"taxonomyID",
			"lastName",
		}
		//
		m0  = db.Model(tbl0)
		dl0 []*Programmer
	)
	as.Nil(m0.Objects().All(&dl0))
	as.Equal(2, len(dl0), "all object")
	//
	dl0 = []*Programmer{}
	as.Nil(m0.Objects().Filter(orm.M{"taxonomyID": "9606"}).Limit(10).All(&dl0))
	as.Equal(2, len(dl0), "filter object")
	//
	dl0 = []*Programmer{}
	as.Nil(m0.ObjectsWith(&orm.ArgObjects{
		LogLevel: orm.LevelDebug,
	}).Group(gps...).Limit(10).All(&dl0))
	as.Equal(1, len(dl0), "group object all")
	//
	dl0 = []*Programmer{}
	as.Nil(m0.ObjectsWith(&orm.ArgObjects{
		LogLevel: orm.LevelDebug,
	}).Group(gps...).Filter(orm.M{"taxonomyID": "9606"}).Sort(gps...).Limit(10).All(&dl0))
	as.Equal(1, len(dl0), "group object filter")
	if true {
		_count, _err := m0.ObjectsWith(&orm.ArgObjects{
			LogLevel: orm.LevelDebug,
		}).Group(gps...).Filter(orm.M{"taxonomyID": "9606"}).Sort(gps...).Limit(10).Count()
		as.Nil(_err)
		as.Equal(1, _count)
	}
}

// 压力测试:建表删表
func Benchmark_ModelCreateDrop(b *testing.B) {
	db := testGetDB()
	b.Logf(`ModelCreateDrop: %v`, db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			tbl0 = fmt.Sprintf("bench_programmer_%d", i)
			m0   = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelError})
			err  error
		)
		if err = m0.Ensure(&Programmer{}); err != nil {
			b.Fatalf("%s err: %v", tbl0, err)
		}
		if err = m0.Drop(); err != nil {
			b.Fatalf("%s err: %v", tbl0, err)
		}
	}
}

// 压测新建数据
func Benchmark_ModelObjects(b *testing.B) {
	db := testGetDB()
	b.Logf(`ModelObjects: %v`, db)

	var (
		tbl0 = "bench_object"
		m0   = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelError})
		err  error
	)
	if err = m0.Drop(); err != nil {
		b.Fatal(err)
	}
	if err = m0.EnsureColumn(&Programmer{}); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := new(Programmer)
		n.UID = fmt.Sprintf("%d", i)
		n.Key = fmt.Sprintf("%d", i)
		n.MetaJSON = `{}`
		n.Birthday = time.Now()
		if err = m0.Objects().Create(n); err != nil {
			b.Fatal(err)
			break
		}
	}
}

// 压测新建数据:带索引
func Benchmark_ModelObjectsIndex(b *testing.B) {
	db := testGetDB()
	b.Logf(`ModelObjectsIndex: %v`, db)

	var (
		tbl0 = "bench_object_index"
		m0   = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelError})
		err  error
	)
	if err = m0.Drop(); err != nil {
		b.Fatal(err)
	}
	if err = m0.Ensure(&Programmer{}); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := new(Programmer)
		n.UID = fmt.Sprintf("%d", i)
		n.Key = fmt.Sprintf("%d", i)
		n.FirstName = n.UID
		n.LastName = n.UID
		n.MetaJSON = `{}`
		n.Birthday = time.Now()
		if err = m0.Objects().Create(n); err != nil {
			b.Fatal(err)
			break
		}
	}
}

// 压测读取数据
func Benchmark_ModelObjectsQuery(b *testing.B) {
	db := testGetDB()
	b.Logf(`ModelObjectsQuery: %v`, db)

	var (
		tbl0 = "bench_object"
		m0   = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelError})
		err  error
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var d []*Programmer
		if err = m0.Objects().Sort("id").Filter(orm.M{
			"uid": fmt.Sprintf("%d", rand.Int())}).Limit(1).All(&d); err != nil {
			b.Fatal(err)
		}
	}
}

// 压测读取数据:带索引
func Benchmark_ModelObjectsQueryIndex(b *testing.B) {
	db := testGetDB()
	b.Logf(`ModelObjectsQueryIndex: %v`, db)

	var (
		tbl0 = "bench_object_index"
		m0   = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelError})
		err  error
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var d []*Programmer
		if err = m0.Objects().Sort("id").Filter(orm.M{
			"uid": fmt.Sprintf("%d", rand.Int())}).Limit(1).All(&d); err != nil {
			b.Fatal(err)
		}
	}
}

// 并发测试读数据:带索引
func Benchmark_ModelObjectsQueryParallel(b *testing.B) {
	db := testGetDB()
	b.Logf(`ModelObjectsQueryParallel: %v`, db)

	var (
		tbl0 = "bench_object_index"
		m0   = db.Model(tbl0).With(&orm.ArgModel{LogLevel: orm.LevelError})
		err  error
	)

	//b.SetParallelism(4)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var d []*Programmer
			if err = m0.Objects().Sort("id").Filter(orm.M{
				"uid": fmt.Sprintf("%d", rand.Int())}).Limit(1).All(&d); err != nil {
				b.Fatal(err)
			}
		}
	})
}
