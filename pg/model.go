package postgres

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"

	"database/sql"
	"fmt"
	"strings"
	"sync"
)

type Model struct {
	TableName string
	sync.RWMutex
	// sql
	DatabaseSql *DatabaseSql
	Result      sql.Result
	// sql contig of objects
	ContigInsert string
	ContigUpdate string
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
		columns = append(columns, f.Name)
		values = append(values, ":"+f.Name)
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
	)
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
		default:
			break
		}

		// pk
		if f.Primary == true {
			if len(primaryKey) > 0 {
				log.Warn("define dunplicatly primary key, use last: ", primaryKey, " ", f.Name)
			}
			primaryKey = f.Name
		}

		if f.Kind == "varchar" && f.Size > 0 {
			colCmdLis = append(colCmdLis, fmt.Sprintf("\"%s\" %s(%d)", f.Name, f.Kind, f.Size))
		} else {
			colCmdLis = append(colCmdLis, fmt.Sprintf("\"%s\" %s", f.Name, f.Kind))
		}
	}

	// primary key
	if len(primaryKey) > 0 {
		colCmdLis = append(colCmdLis, fmt.Sprintf("CONSTRAINT %s_pkey PRIMARY KEY (%s)", m.TableName, primaryKey))
	}

	// combine sql
	colCmd := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n%s\n);", m.TableName, strings.Join(colCmdLis, ",\n"))

	//log.Debug(colCmd)
	if m.Result, err = m.DatabaseSql.DB.Exec(colCmd); err != nil {
		log.Error("sql: ", colCmd, " error: ", err.Error())
		return
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

// 事务
func (m *Model) Begin() (orm.Trans, error) {
	var (
		t   = &Trans{}
		err error
	)
	if t.Tx, err = m.DatabaseSql.DB.Beginx(); err != nil {
		return t, err
	}
	return t, err
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
