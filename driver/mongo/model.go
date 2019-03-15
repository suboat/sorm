package mongo

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"

	//"gopkg.in/mgo.v2"
	"github.com/globalsign/mgo"

	"sync"
)

// Model model
type Model struct {
	sync.RWMutex
	//
	TableName  string
	Collection *mgo.Collection // mgo table
	//
	log orm.Logger
}

// Copy 全拷贝
func (m *Model) Copy() (r *Model) {
	r = new(Model)
	r.TableName = m.TableName
	r.Collection = m.Collection
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

func (m *Model) String() string {
	return m.TableName
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
	if m.Collection != nil {
		if err = m.Collection.DropCollection(); err != nil && err.Error() == "ns not found" {
			err = nil
		}
	}
	return
}

// EnsureIndex 确保索引存在
func (m *Model) EnsureIndex(indexMap orm.Index) (err error) {
	// default attr
	index := mgo.Index{
		Key:        []string{},
		Unique:     false,
		DropDups:   false, // important
		Background: true,
		Sparse:     true,
	}
	// keys
	if i, ok := indexMap["Key"]; ok {
		if v, ok := i.([]string); ok {
			index.Key = v
		}
	}
	// unique
	if i, ok := indexMap["Unique"]; ok {
		if v, ok := i.(bool); ok {
			index.Unique = v
		}
	}
	// DropDups
	if i, ok := indexMap["DropDups"]; ok {
		if v, ok := i.(bool); ok {
			index.DropDups = v
		}
	}
	// DefaultLanguage
	if i, ok := indexMap["DefaultLanguage"]; ok {
		if v, ok := i.(string); ok {
			index.DefaultLanguage = v
		}
	}
	// LanguageOverride
	if i, ok := indexMap["LanguageOverride"]; ok {
		if v, ok := i.(string); ok {
			index.LanguageOverride = v
		}
	}
	// Sparse
	if i, ok := indexMap["Sparse"]; ok {
		if v, ok := i.(bool); ok {
			index.Sparse = v
		}
	}
	err = m.Collection.EnsureIndex(index)
	return
}

// EnsureColumn 确认字段
func (m *Model) EnsureColumn(st interface{}) (err error) {
	return m.Ensure(st)
}

// Ensure is sync table with struct
func (m *Model) Ensure(st interface{}) (err error) {
	// index
	var (
		fieldInfoLis []*orm.FieldInfo
	)
	if fieldInfoLis, err = orm.StructModelInfo(st); err != nil {
		return
	}
	for _, f := range fieldInfoLis {
		if !f.Index && !f.Unique && !f.IndexText && !f.Primary {
			continue
		}
		if f.Index {
			if err = m.EnsureIndex(orm.Index{
				"Key":      f.IndexKeys,
				"Unique":   false,
				"DropDups": false, // mongodb
			}); err != nil {
				return
			}
		}
		if f.Unique {
			if err = m.EnsureIndex(orm.Index{
				"Key":      f.UniqueKeys,
				"Unique":   true,
				"DropDups": true, // mongodb:important
			}); err != nil {
				return
			}
		}
		// fallback to unique
		if f.Primary && !f.Serial && f.Name != "id" {
			if err = m.EnsureIndex(orm.Index{
				"Key":      f.PrimaryKeys,
				"Unique":   true,
				"DropDups": true, // mongodb:important
			}); err != nil {
				return
			}
		}
		if f.IndexText {
			if len(f.IndexKeys) == 1 {
				if err = m.EnsureIndex(orm.Index{
					"Key":              []string{"$text:" + f.IndexKeys[0]},
					"DefaultLanguage":  "en",
					"LanguageOverride": "en",
				}); err != nil {
					return
				}
			} else {
				err = orm.ErrIndexTextParamsInvalid
				return
			}
		}
	}
	return
}

// Begin 事务起始
func (m *Model) Begin() (orm.Trans, error) {
	return m.BeginWith(nil)
}

// BeginWith 事务
func (m *Model) BeginWith(arg *orm.ArgTrans) (ret orm.Trans, err error) {
	if CfgTxUnsafe {
		return &Trans{}, nil
	} else {
		return nil, orm.ErrTransNotSupport
	}
}

// Rollback 事务回滚
func (m *Model) Rollback(t orm.Trans) error {
	if CfgTxUnsafe {
		return t.Rollback()
	} else {
		return orm.ErrTransNotSupport
	}
}

// Commit 事务提交
func (m *Model) Commit(t orm.Trans) error {
	if CfgTxUnsafe {
		return t.Commit()
	} else {
		return orm.ErrTransNotSupport
	}
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

// Exec 兼容sql
func (m *Model) Exec(query string, args ...interface{}) (result orm.Result, err error) {
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
