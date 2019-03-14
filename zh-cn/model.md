# 模型

> `模型`是基于源数据库的抽象实例。模型包含基本字段和操作数据的方法。  
> 简而言之，一个`模型`是一个`数据库表`的映射。

# 快速入门

例子中的结构体`DemoAnimal`定义了`domain`、`height`、`weight`、`length`字段，
并忽略了`sound`字段:

```go
// 模型定义
type DemoAnimal struct {
	Domain string  `sorm:"unique;size(36)"`
	Height float32 `sorm:"index"`
	Weight float32
	Length int64
	Sound  string `sorm:"-"`
}

// 自动建表
func demoModelEnsureIndexWithTag(db orm.Database) {
	if err := db.Model("demoanimal").EnsureIndexWithTag(&DemoAnimal{}); err != nil {
		panic(err)
	}
}
```

与此同时，`domain`标记了`unique`属性，`height`标记了`index`属性。DemoAnimal`模型`将会产生如下`sql`语句:

```sql
-- postgres9.6
CREATE TABLE demoperson (
    "domain" varchar(36),
    "height" float,
    "weight" float,
    "length" bigint
);
CREATE UNIQUE INDEX "demoperson_domain" ON "demoperson"  ("domain");
CREATE INDEX "demoperson_height" ON "demoperson"  ("height");
```

# 进阶入门

模型间的组合，如例子中的`DemoAnimal`、`demoPerson`、`demoAddress`结构体组合:

```go
type DemoAnimal struct {
	Domain string  `sorm:"unique;size(36)"`
	Height float32 `sorm:"index"`
	Weight float32
	Length int64
	Sound  string `sorm:"-"`
}

type demoPerson struct {
	Uid         string           `sorm:"size(36) primary"`
	*DemoAnimal `bson:",inline"` // mgo default embed fields

	FirstName string         `sorm:"size(32);index"`
	LastName  string         `sorm:"size(64);unique(firstname)"`
	Age       int            `sorm:"index"`      // 数字类型
	Birthday  time.Time      `sorm:"index"`      // 时间戳类型
	Message   string         `sorm:"json;index"` // 主动声明为json类型
	Address   demoAddress    ``                  // 内嵌类型
	Password  []byte         ``                  // 二进制类型
	Tags      types.SliceInf `sorm:"index"`      // 数组类型
}

type demoAddress struct {
	Road   string
	Number int
}

// 自动建表
func demoModelEnsureIndexWithTag(db orm.Database) {
	var (
		m   = db.Model("demoperson")
		p   = &demoPerson{DemoAnimal: &DemoAnimal{}}
		err error
	)
	if err = m.EnsureIndexWithTag(p); err != nil {
		panic(err)
	}
}
```

对应的`sql`语句:

```sql
-- postgres9.6
CREATE TABLE demoperson (
    "uid" varchar(36),
    "domain" varchar(36),
    "height" float,
    "weight" float,
    "length" bigint,
    "firstname" varchar(32),
    "lastname" varchar(64),
    "age" integer,
    "birthday" timestamp with time zone,
    "message" jsonb,
    "address" jsonb,
    "password" bytea,
    "tags" jsonb,
    CONSTRAINT demoperson_pkey PRIMARY KEY ("uid")
);
CREATE UNIQUE INDEX "demoperson_domain" ON "demoperson"  ("domain");
CREATE INDEX "demoperson_height" ON "demoperson"  ("height");
CREATE INDEX "demoperson_firstname" ON "demoperson"  ("firstname");
CREATE UNIQUE INDEX "demoperson_lastname_firstname" ON "demoperson"  ("lastname", "firstname");
CREATE INDEX "demoperson_age" ON "demoperson"  ("age");
CREATE INDEX "demoperson_birthday" ON "demoperson"  ("birthday");
CREATE INDEX "demoperson_message" ON "demoperson" USING gin ("message");
CREATE INDEX "demoperson_tags" ON "demoperson" USING gin ("tags");
```

此外，可以对定义过的模型做追加定义，`orm`会自动执行`ALTER TABLE`等语句。  
无法覆盖定义，只能删除表字段后再同步或手动修改数据库。

# 定义模型

> 通过结构体的`tag`来定义字段属性，关键词`sorm`。  
> 为了兼容更多的数据库，字段名全为小写。

|   Key   |       Args        | Example                  |          Comment           |
|:-------:|:-----------------:|:-------------------------|:--------------------------:|
| primary |         -         | sorm:"primary"           |         设置为主键          |
| serial  |         -         | sorm:"serial"            |         自增数字主键         |
| unique  | relative field(s) | sorm:"unique(uid,email)" |         (联合)唯一          |
|  index  | relative field(s) | sorm:"index(uid,email)"  |         (联合)索引          |
|  json   |         -         | sorm:"json"              | 强制以数据库支持的json格式储存 |
|  size   |   varchar size    | sorm:"size(36)"          |        以36长度储存         |
|    -    |         -         | bson:",inline"           |    兼容`mgo`的结构体组合     |

# 模型方法

## Objects(对象)
> 返回`Objects`类型，可以执行`创建`、`删除`、`更新`、[查询](zh-cn/query.md?id=查询数据)等操作。如:

```go

// create
func demoCreateData(m orm.Model) (ret *demoPerson) {
	ret = new(demoPerson)
	ret.Age = 16
	if err = m.Objects().Create(ret); err != nil {
		panic(err)
	}
	return
}

// filter
func demoGetData(m orm.Model) (ret []*demoPerson) {
	ret = []*demoPerson{}
	if err = m.Objects().Filter(orm.M{
		"age": 16,
	}).All(&ret); err != nil {
		panic(err)
	}
	return
}

// update
func demoUpdateData(m orm.Model) (err error) {
	if err = m.Objects().Filter(orm.M{
		"age": 16,
	}).Update(map[string]interface{}{
		"age": 18,
	}); err != nil {
		return
	}
	return
}

// delete
func demoDeleteData(m orm.Model) (err error) {
	if err = m.Objects().Filter(orm.M{
		"age": 18,
	}).Delete(); err != nil {
		return
	}
	return
}

```

## Begin(事务)
> 返回`Trans`类型，提供事务操作。

```go
// 事务使用
func demoTrans(m orm.Model) (err error) {
	var (
		query = orm.M{"sid": "12345678"}
		tx    orm.Trans
	)

	// 事务开始
	if tx, err = m.Begin(); err != nil {
		return
	} else {
		// 将函数内的错误告知事务
		defer func() {
			if err != nil {
				tx.ErrorSet(err)
			}
			// 自动处理事务的 commit/rollback
			m.AutoTrans(tx)
		}()
	}

	// 事务内行锁
	if err = m.Objects().Filter(query).TLockUpdate(tx); err != nil {
		return
	}

	// amount自增1
	if err = m.Objects().Filter(query).TUpdate(map[string]interface{}{
		orm.TagUpdateInc: map[string]interface{}{
			"amount": 1,
		},
	}, tx); err != nil {
		return
	}

	return
}
```