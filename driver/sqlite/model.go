package sqlite

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

// Model 实现
type Model struct {
	TableName string
	sync.RWMutex
	// sql
	DatabaseSQL *DatabaseSQL
	Result      sql.Result
	// sql contig of objects
	ContigInsert       string
	ContigUpdate       string
	AutoIncrementField string
	//
	log orm.Logger
}

// SQLIndex is index info of SQL-like database
type SQLIndex struct {
	Key    []string
	Unique bool
	Kind   string // field type
	Method string // 索引方法
}

// Copy 全拷贝
func (m *Model) Copy() (r *Model) {
	r = new(Model)
	r.TableName = m.TableName
	r.DatabaseSQL = m.DatabaseSQL
	r.ContigInsert = m.ContigInsert
	r.ContigUpdate = m.ContigUpdate
	r.AutoIncrementField = m.AutoIncrementField
	if m.log != nil {
		if _log, ok := m.log.(*log.Logger); ok {
			r.log = _log.Copy()
		}
	}
	if r.log == nil {
		if r.log = orm.Log; r.log == nil {
			r.log = log.Log.Copy()
		}
	}
	return
}

// table name
func (m *Model) String() (s string) {
	s = m.TableName
	return
}

// Objects 取Objects
func (m *Model) Objects() orm.Objects {
	return m.ObjectsWith(nil)
}

// ObjectsWith 带参数取Objects
func (m *Model) ObjectsWith(arg *orm.ArgObjects) orm.Objects {
	ob := new(Objects)
	ob.Model = m
	ob.count = -1
	ob.nums = -1
	if m.log != nil {
		if _log, ok := m.log.(*log.Logger); ok {
			ob.log = _log.Copy()
		}
	}
	if ob.log == nil {
		if ob.log = orm.Log; ob.log == nil {
			ob.log = log.Log.Copy()
		}
	}
	if arg != nil {
		if arg.LogLevel > 0 {
			ob.log.SetLevel(arg.LogLevel)
		}
	}
	return ob
}

// Drop table
func (m *Model) Drop() (err error) {
	m.Result, err = m.DatabaseSQL.DB.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, m.TableName))
	return
}

// ContigParse 将结构体信息解析后缓存SQL语句块
func (m *Model) ContigParse(st interface{}) (err error) {
	if len(m.ContigInsert) > 0 && len(m.ContigUpdate) > 0 {
		return
	}
	var (
		columns      []string
		values       []string
		fieldInfoLis []*orm.FieldInfo
	)
	if fieldInfoLis, err = orm.StructModelInfo(st); err != nil {
		return
	}
	for _, f := range fieldInfoLis {
		if f.Name != m.AutoIncrementField {
			columns = append(columns, "\""+f.Name+"\"")
			values = append(values, ":"+f.Name)
		}
	}
	m.ContigInsert = fmt.Sprintf(`"%s" (%s) VALUES (%s)`, m.TableName, strings.Join(columns, ", "), strings.Join(values, ", "))
	m.ContigUpdate = fmt.Sprintf(`"%s" SET (%s) = (%s)`, m.TableName, strings.Join(columns, ", "), strings.Join(values, ", "))
	return
}

// EnsureColumn 确认字段
func (m *Model) EnsureColumn(st interface{}) (err error) {
	var (
		fieldInfoLis []*orm.FieldInfo
		colCmdLis    = []string{}
		primaryKey   = ""
		primaryCmd   = ""
		colCmd       = ""
		tableExist   = 0
		fieldExist   = make(map[string]bool)
	)
	// table exist
	if err = m.DatabaseSQL.DB.Get(&tableExist,
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name=?`, m.TableName); err != nil {
		m.log.Errorf("[ensure-column] check table exist err: %v", err)
		return
	}

	if tableExist == 1 {
		// test column
		var (
			columnLis []string
			query     string
		)
		query = "SELECT * FROM " + m.TableName
		if rows, _err := m.DatabaseSQL.DB.Query(query); _err == nil {
			columnLis, _ = rows.Columns()
		} else {
			err = _err
			m.log.Errorf("[ensure-column] select column err: %v", err)
			return
		}

		for _, col := range columnLis {
			fieldExist[col] = true
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
		case "json":
			// default use jsonb
			f.Kind = "json"
		case "bytearray":
			f.Kind = "bytea"
		case "timestamp":
			// use timezone
			f.Kind = "datetime"
		case "serial", "bigserial":
			if len(m.AutoIncrementField) > 0 && m.AutoIncrementField != f.Name {
				m.log.Errorf(`[ensure-column] dunplicatly AutoIncrementField, last: '%s' now: '%s'`,
					m.AutoIncrementField, f.Name)
				return
			}
			m.AutoIncrementField = f.Name
		default:
		}

		// pk
		if f.Primary {
			if len(primaryKey) > 0 {
				//log.Warn("define dunplicatly primary key, use last: ", primaryKey, " ", f.Name)
				err = fmt.Errorf("define dunplicatly primary key, use last: %s %s", primaryKey, f.Name)
				return
			}
			primaryKey = strings.Join(f.PrimaryKeys, "_")
			// primary keys
			keys := []string{}
			for _, k := range f.PrimaryKeys {
				keys = append(keys, "\""+k+"\"")
			}
			if tableExist == 1 {
				primaryCmd = fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s)", m.TableName, strings.Join(keys, ", "))
			} else {
				primaryCmd = fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(keys, ", "))
			}
		}

		// col
		if tableExist == 1 {
			// add
			if _, ok := fieldExist[f.Name]; ok {
				continue
			}
			colCmd := fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" %s DEFAULT '%v';`, m.TableName, f.Name, f.Kind, f.DefaultVal)
			switch f.Kind {
			case "serial", "bigserial":
				// not default
				colCmd = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" INTEGER DEFAULT 0;`, m.TableName, f.Name)
			case "varchar", "char":
				if f.Size > 0 {
					colCmd = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" %s(%d) DEFAULT '%v';`, m.TableName,
						f.Name, f.Kind, f.Size, f.DefaultVal)
				}
			case "decimal", "numeric":
				if f.Size > 0 {
					if f.Precision >= 0 {
						colCmd = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" %s (%d,%d) DEFAULT %v;`, m.TableName,
							f.Name, f.Kind, f.Size, f.Precision, f.DefaultVal)
					} else {
						colCmd = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" %s (%d) DEFAULT %v;`, m.TableName,
							f.Name, f.Kind, f.Size, f.DefaultVal)
					}
				} else {
					colCmd = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" float DEFAULT %v;`, m.TableName, f.Name, f.DefaultVal)
				}
			case "int", "float":
				colCmd = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" %s DEFAULT %v;`, m.TableName, f.Name, f.Kind, f.DefaultVal)
			default:

			}
			colCmdLis = append(colCmdLis, colCmd)
		} else {
			// create
			colCmd := fmt.Sprintf(`"%s" %s`, f.Name, f.Kind)
			switch f.Kind {
			case "varchar", "char":
				if f.Size > 0 {
					colCmd = fmt.Sprintf(`"%s" %s(%d)`, f.Name, f.Kind, f.Size)
				}
			case "decimal", "numeric":
				if f.Size > 0 {
					if f.Precision >= 0 {
						colCmd = fmt.Sprintf(`"%s" %s (%d,%d)`, f.Name, f.Kind, f.Size, f.Precision)
					} else {
						colCmd = fmt.Sprintf(`"%s" %s (%d)`, f.Name, f.Kind, f.Size)
					}
				} else {
					colCmd = fmt.Sprintf(`"%s" float`, f.Name)
				}
			case "serial", "bigserial":
				colCmd = fmt.Sprintf(`"%s" INTEGER DEFAULT 0`, f.Name)
			default:
				break
			}
			colCmdLis = append(colCmdLis, colCmd)
		}
	}

	// new or add
	if tableExist == 1 {
		// exist
		//colCmd = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n%s\n);", m.TableName, strings.Join(colCmdLis, ",\n"))
		colCmd = fmt.Sprintf("%s\n;", strings.Join(colCmdLis, "\n"))
	} else {
		// primary key
		if len(primaryCmd) > 0 {
			colCmdLis = append(colCmdLis, primaryCmd)
		}
		colCmd = fmt.Sprintf("CREATE TABLE \"%s\" (\n%s\n);", m.TableName, strings.Join(colCmdLis, ",\n"))
	}

	// exec
	if len(colCmdLis) > 0 {
		if m.Result, err = m.DatabaseSQL.DB.Exec(colCmd); err != nil {
			m.log.Errorf(`[ensure-column] 
%s err: %v`, colCmd, err)
			return
		}

		// log
		m.log.Infof(`[ensure-column] 
%s`, colCmd)
	}

	return
}

// EnsurePrimary 确认主键(sqlite不支持建表后添加重要约束)
func (m *Model) EnsurePrimary(key []string) (err error) {
	//	var (
	//		pkey = fmt.Sprintf(`%s_pkey`, m.TableName)
	//		keys []string
	//		cmd  string
	//	)
	//	for _, k := range key {
	//		keys = append(keys, "\""+strings.ToLower(k)+"\"")
	//	}
	//
	//	// cmd
	//	cmd = fmt.Sprintf(`DO $$
	//BEGIN
	//    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = '%s') THEN
	//        %s
	//    END IF;
	//END;$$;`, pkey, fmt.Sprintf(
	//		`ALTER TABLE "%s" ADD CONSTRAINT "%s" PRIMARY KEY(%s);`,
	//		m.TableName, pkey, strings.Join(keys, ", ")))
	//
	//	// run
	//	if m.Result, err = m.DatabaseSQL.DB.Exec(cmd); err != nil {
	//		m.log.Errorf(`[ensure-primary] "%s" err: %v`, cmd, err)
	//		return
	//	}
	return
}

// EnsureIndex 确保索引存在
func (m *Model) EnsureIndex(indexMap orm.Index) (err error) {
	var (
		indexLis       []*SQLIndex // index and unique
		indexMethod    = ""        // both index and unique method
		indexFiledType = ""        // field type
	)

	// filed type
	if s, ok := indexMap["Kind"].(string); ok {
		indexFiledType = s
	}

	// method, like json
	if i, ok := indexMap["Method"]; ok {
		if s, ok := i.(string); ok {
			indexMethod = s
		}
	}

	// keys
	if i, ok := indexMap["Key"]; ok {
		if v, ok := i.([]string); ok {
			indexLis = append(indexLis, &SQLIndex{
				Key:    v,
				Unique: false,
				Kind:   indexFiledType,
				Method: indexMethod,
			})
		}
	}

	// unique
	if i, ok := indexMap["Unique"]; ok {
		if v, ok := i.([]string); ok {
			indexLis = append(indexLis, &SQLIndex{
				Key:    v,
				Unique: true,
				Kind:   indexFiledType,
				Method: indexMethod,
			})
		}
	}

	//
	for _, index := range indexLis {
		// cmd
		if len(index.Key) > 0 {
			indexType := "INDEX"
			if index.Unique {
				indexType = "UNIQUE INDEX"
			}
			var keys []string
			for _, k := range index.Key {
				k = strings.ToLower(k)
				keys = append(keys, "\""+k+"\"")
			}
			//if len(index.Method) == 0 {
			//	index.Method = "USING btree" // 默认 btree
			//}
			indexCmd := fmt.Sprintf("CREATE %s IF NOT EXISTS \"%s_%s\" ON \"%s\" %s (%s);", indexType,
				m.TableName, strings.Join(index.Key, "_"), m.TableName, index.Method, strings.Join(keys, ", "))

			if m.Result, err = m.DatabaseSQL.DB.Exec(indexCmd); err != nil {
				m.log.Errorf(`[ensure-index] %s, err: %v`, indexCmd, err)
				return

			}
			m.log.Debug(indexCmd)
		}
	}

	return
}

// Ensure is sync table with struct
func (m *Model) Ensure(st interface{}) (err error) {
	// syncDB
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
		if f.Primary {
			// primary
			//if err = m.EnsurePrimary(f.PrimaryKeys); err != nil {
			//	return
			//}
			continue
		} else if !f.Index && !f.Unique {
			// ignore
			continue
		}

		idxM := map[string]interface{}{"Kind": f.Kind}
		if f.Index {
			idxM["Key"] = f.IndexKeys
		}
		if f.Unique {
			idxM["Unique"] = f.UniqueKeys
		}
		//if (f.Kind == "json") && (f.Index) {
		//	idxM["Method"] = "USING gin" // for jsonb
		//}
		if err = m.EnsureIndex(orm.Index(idxM)); err != nil {
			return
		}
	}
	// contig parse and cache
	m.ContigInsert = ""
	m.ContigUpdate = ""
	if err = m.ContigParse(st); err != nil {
		return
	}
	return
}

// Begin 事务起始
func (m *Model) Begin() (orm.Trans, error) {
	return m.BeginWith(nil)
}

// BeginWith 事务
func (m *Model) BeginWith(arg *orm.ArgTrans) (ret orm.Trans, err error) {
	var (
		t = &Trans{}
	)
	if t.Tx, err = m.DatabaseSQL.DB.Beginx(); err != nil {
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

		_ = t.DebugPush(fmt.Sprintf("%s:%d|%s", filePath, line, fnName))
	} else {
		_ = t.DebugPush("can not alloc call path")
	}

	if arg == nil {
		return t, nil
	}

	// 事务级别
	if arg != nil {
		if len(arg.Level) > 0 {
			switch arg.Level {
			case orm.TransReadUncommitted:
				if _, err = t.Exec("PRAGMA read_uncommitted = true;"); err != nil {
					return
				}
			case orm.TransReadCommitted:
				if _, err = t.Exec("PRAGMA read_committed = true;"); err != nil {
					return
				}
			case orm.TransRepeatableRead:
				if _, err = t.Exec("PRAGMA repeatable_read = true;"); err != nil {
					return
				}
			case orm.TransSerializable:
				if _, err = t.Exec("PRAGMA serializable = true;"); err != nil {
					return
				}
			default:
				err = orm.ErrTransLevelUnknown
				_ = t.Rollback()
				return
			}
		}
	}

	// 超时回滚
	t.timer = time.NewTimer(orm.TransTimeout)
	go func() {
		defer func() {
			if _err := recover(); _err != nil {
				m.log.Error(err)
			}
		}()

		<-t.timer.C
		// timeout
		if !t.isFinish {
			// log sql history
			m.log.Errorf("[trans-timeout] tableName=%s sql=%s", m.TableName, t.debugReport())
			_ = m.Rollback(t)
		}
	}()

	return t, nil
}

// Rollback 事务回滚
func (m *Model) Rollback(t orm.Trans) error {
	if t == nil {
		return orm.ErrTransEmpty
	}
	return t.Rollback()
}

// Commit 事务提交
func (m *Model) Commit(t orm.Trans) error {
	if t == nil {
		return orm.ErrTransEmpty
	}
	return t.Commit()
}

// AutoTrans 自动处理事务
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

// Exec 执行语句
func (m *Model) Exec(query string, args ...interface{}) (result orm.Result, err error) {
	result, err = m.DatabaseSQL.DB.Exec(query, args...)
	return
}

// Exec 执行语句
func (m *Model) Select(dest interface{}, query string, args ...interface{}) (err error) {
	err = m.DatabaseSQL.DB.Select(dest, query, args...)
	return
}

// With 设置日志级别
func (m *Model) With(arg *orm.ArgModel) (ret orm.Model) {
	r := m.Copy()
	if arg != nil {
		if arg.LogLevel > 0 {
			r.log.SetLevel(arg.LogLevel)
		}
	}
	return r
}
