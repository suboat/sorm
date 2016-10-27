package postgres

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"

	"database/sql"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Model struct {
	TableName string
	sync.RWMutex
	// sql
	DatabaseSql *DatabaseSql
	Result      sql.Result
	// sql contig of objects
	ContigInsert       string
	ContigUpdate       string
	AutoIncrementField string
}

// index info of SQL-like database
type SqlIndex struct {
	Key    []string
	Unique bool
}

// table name
func (m *Model) String() (s string) {
	s = m.TableName
	return
}

// objects
func (m *Model) Objects() orm.Objects {
	ob := new(Objects)
	ob.Model = m
	return ob
}

// object parse
// sql contig functions
func (m *Model) ContigParse(st interface{}) (err error) {
	if len(m.ContigInsert) > 0 && len(m.ContigUpdate) > 0 {
		return
	}
	var (
		columns      = []string{}
		values       = []string{}
		fieldInfoLis []*orm.FieldInfo
	)
	if fieldInfoLis, err = orm.StructModelInfo(st); err != nil {
		return
	}
	for _, f := range fieldInfoLis {
		if f.Name != m.AutoIncrementField {
			columns = append(columns, f.Name)
			values = append(values, ":"+f.Name)
		}
	}
	m.ContigInsert = fmt.Sprintf("%s (%s) VALUES (%s)", m.TableName, strings.Join(columns, ", "), strings.Join(values, ", "))
	m.ContigUpdate = fmt.Sprintf("%s SET (%s) = (%s)", m.TableName, strings.Join(columns, ", "), strings.Join(values, ", "))
	return
}

// uid
func (m *Model) NewUid() orm.Uid {
	return orm.NewUid() // 采用默认生成
}

// valid
func (m *Model) UidValid(u orm.Uid) (err error) {
	return u.Valid()
}

// ensure column: create table
func (m *Model) EnsureColumn(st interface{}) (err error) {
	var (
		fieldInfoLis []*orm.FieldInfo
		colCmdLis    = []string{}
		primaryKey   = ""
		colCmd       = ""
		tableExist   = 0
		fieldExist   = make(map[string]bool)
	)
	// table exist
	if err = m.DatabaseSql.DB.Get(&tableExist,
		"SELECT count(*) FROM information_schema.tables WHERE table_name=$1", m.TableName); err != nil {
		log.Error("check table exist ", err)
		return
	}
	if tableExist == 1 {
		// test column
		var columnLis = []string{}
		if err = m.DatabaseSql.DB.Select(&columnLis,
			"SELECT column_name FROM information_schema.columns WHERE table_name=$1", m.TableName); err != nil {
			log.Error("select column ", err)
			return
		} else {
			for _, col := range columnLis {
				fieldExist[col] = true
			}
		}
	}
	// fields parser
	if fieldInfoLis, err = orm.StructModelInfo(st); err != nil {
		return
	}
	for _, f := range fieldInfoLis {
		switch f.Kind {
		case "unit":
			f.Kind = "int"
			break
		case "json":
			// default use jsonb
			f.Kind = "jsonb"
			break
		case "bytearray":
			f.Kind = "bytea"
			break
		case "timestamp":
			// use timezone
			f.Kind = "timestamp with time zone"
			break
		case "serial", "bigserial":
			if len(m.AutoIncrementField) > 0 {
				err = fmt.Errorf("dunplicatly AutoIncrementField, last: '%s' now: '%s'", m.AutoIncrementField, f.Name)
				return
			}
			m.AutoIncrementField = f.Name
			break
		default:
			break
		}

		// pk
		if f.Primary == true {
			if len(primaryKey) > 0 {
				//log.Warn("define dunplicatly primary key, use last: ", primaryKey, " ", f.Name)
				err = fmt.Errorf("define dunplicatly primary key, use last: %s %s", primaryKey, f.Name)
				return
			}
			primaryKey = f.Name
		}

		if tableExist == 1 {
			if _, ok := fieldExist[f.Name]; ok == true {
				continue
			}
			if f.Kind == "varchar" && f.Size > 0 {
				colCmdLis = append(colCmdLis, fmt.Sprintf(`ADD COLUMN "%s" %s(%d) DEFAULT '%v'`,
					f.Name, f.Kind, f.Size, f.DefaultVal))
			} else {
				colCmdLis = append(colCmdLis, fmt.Sprintf(`ADD COLUMN "%s" %s DEFAULT '%v'`,
					f.Name, f.Kind, f.DefaultVal))
			}
		} else {
			if f.Kind == "varchar" && f.Size > 0 {
				colCmdLis = append(colCmdLis, fmt.Sprintf("\"%s\" %s(%d)", f.Name, f.Kind, f.Size))
			} else {
				colCmdLis = append(colCmdLis, fmt.Sprintf("\"%s\" %s", f.Name, f.Kind))
			}
		}
	}

	// new or add
	if tableExist == 1 {
		// exist
		//colCmd = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n%s\n);", m.TableName, strings.Join(colCmdLis, ",\n"))
		colCmd = fmt.Sprintf("ALTER TABLE %s \n%s\n;", m.TableName, strings.Join(colCmdLis, ",\n"))
	} else {
		// primary key
		if len(primaryKey) > 0 {
			colCmdLis = append(colCmdLis, fmt.Sprintf("CONSTRAINT %s_pkey PRIMARY KEY (%s)", m.TableName, primaryKey))
		}
		colCmd = fmt.Sprintf("CREATE TABLE %s (\n%s\n);", m.TableName, strings.Join(colCmdLis, ",\n"))
	}

	//log.Debug(colCmd)
	//println(colCmd)
	if len(colCmdLis) > 0 {
		if m.Result, err = m.DatabaseSql.DB.Exec(colCmd); err != nil {
			log.Error("sql: ", colCmd, " error: ", err.Error())
			return
		}
	}

	return
}

// index
func (m *Model) EnsureIndex(indexMap orm.Index) (err error) {
	// default index
	index := &SqlIndex{
		Key:    []string{},
		Unique: false,
	}
	// keys
	if i, ok := indexMap["Key"]; ok == true {
		if v, ok := i.([]string); ok {
			index.Key = v
		}
	}
	// unique
	if i, ok := indexMap["Unique"]; ok == true {
		if v, ok := i.(bool); ok {
			index.Unique = v
		}
	}
	// cmd
	if len(index.Key) > 0 {
		indexType := "INDEX"
		if index.Unique == true {
			indexType = "UNIQUE INDEX"
		}
		keys := []string{}
		for _, k := range index.Key {
			keys = append(keys, "\""+k+"\"")
		}
		indexCmd := fmt.Sprintf("CREATE %s \"%s_%s\" ON \"%s\" (%s);", indexType,
			m.TableName, strings.Join(index.Key, "_"), m.TableName, strings.Join(keys, ", "))
		//log.Debug(indexCmd)
		if m.Result, err = m.DatabaseSql.DB.Exec(indexCmd); err != nil {
			// ignore exist error
			if strings.Index(err.Error(), "already exists") > -1 {
				err = nil
			} else {
				log.Error(err, " [", indexCmd, "]")
				return
			}
		} else {
			//log.Info(indexCmd)
		}
	}
	return
}

// index auto
func (m *Model) EnsureIndexWithTag(st interface{}) (err error) {
	// syncdb
	if err = m.EnsureColumn(st); err != nil {
		return
	}
	// index
	var (
		fieldInfoLis []*orm.FieldInfo
	)
	if fieldInfoLis, err = orm.StructModelInfo(st); err != nil {
		return
	}
	for _, f := range fieldInfoLis {
		if f.Index == false && f.Unique == false {
			continue
		}
		if f.Index == true {
			if err = m.EnsureIndex(orm.Index{
				"Key":    f.IndexKeys,
				"Unique": false,
			}); err != nil {
				return
			}
		}
		if f.Unique == true {
			if err = m.EnsureIndex(orm.Index{
				"Key":    f.UniqueKeys,
				"Unique": true,
			}); err != nil {
				return
			}
		}
	}
	// contig parse and cache
	if err = m.ContigParse(st); err != nil {
		return
	}
	return
}

// **事务
// 事务起始
func (m *Model) Begin() (orm.Trans, error) {
	var (
		t   = &Trans{}
		err error
	)
	if t.Tx, err = m.DatabaseSql.DB.Beginx(); err != nil {
		return t, err
	}

	// 记录调用处
	if pc, file, line, ok := runtime.Caller(2); ok {
		// func
		_fnName := runtime.FuncForPC(pc).Name()
		_fnNameLis := strings.Split(_fnName, ".")
		_fnNameSrc := strings.Split(_fnName, "/")[0]
		fnName := _fnNameLis[len(_fnNameLis)-1]

		// file
		_pcLis := strings.Split(file, _fnNameSrc)
		filePath := _fnNameSrc + strings.Join(_pcLis[1:], "")

		t.DebugPush(fmt.Sprintf("%s:%d|%s", filePath, line, fnName))
	} else {
		t.DebugPush("can not alloc call path")
	}

	// 超时回滚
	t.timer = time.NewTimer(orm.TransTimeout)
	go func() {
		defer func() {
			if _err := recover(); _err != nil {
				log.Error(err)
			}
		}()

		<-t.timer.C
		// timeout
		if t.isFinish == false {
			// log sql history
			var report = fmt.Sprintf("[TRANS-TIMEOUT] tableName=%s sql=%s", m.TableName, t.debugReport())
			log.Error(report)
			m.Rollback(t)
		}
	}()

	return t, err
}

// 声明级别的事务
func (m *Model) BeginLevel(level string) (t orm.Trans, err error) {
	if t, err = m.Begin(); err != nil {
		return
	}
	switch level {
	case orm.TransReadUncommitted:
		if _, err = t.Exec("set transaction isolation level read uncommitted"); err != nil {
			return
		}
		break
	case orm.TransReadCommitted:
		if _, err = t.Exec("set transaction isolation level read committed"); err != nil {
			return
		}
		break
	case orm.TransRepeatableRead:
		if _, err = t.Exec("set transaction isolation level repeatable read"); err != nil {
			return
		}
		break
	case orm.TransSerializable:
		if _, err = t.Exec("set transaction isolation level serializable"); err != nil {
			return
		}
		break
	default:
		err = orm.ErrTransLevelUnknown
		t.Rollback()
		return
		break
	}
	return
}

func (m *Model) Rollback(t orm.Trans) error {
	if t == nil {
		return orm.ErrTransEmpty
	}
	return t.Rollback()
}
func (m *Model) Commit(t orm.Trans) error {
	if t == nil {
		return orm.ErrTransEmpty
	}
	return nil
}
func (m *Model) AutoTrans(t orm.Trans) (err error) {
	if t == nil {
		err = orm.ErrTransEmpty
		return
	}
	if t.Error() != nil {
		// rollback
		err = t.Rollback()
	} else {
		// commit
		err = t.Commit()
	}
	return
}

// 执行语句
func (m *Model) Exec(query string, args ...interface{}) (result orm.Result, err error) {
	result, err = m.DatabaseSql.DB.Exec(query, args...)
	return
}
