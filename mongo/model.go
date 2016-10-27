package mongo

import (
	"github.com/suboat/sorm"

	"gopkg.in/mgo.v2"
	"sync"
)

type model struct {
	name string
	sync.RWMutex
	collection *mgo.Collection // mgo table
}

func (m *model) String() string {
	return m.name
}

func (m *model) Objects() orm.Objects {
	ob, _ := newObject(nil)
	ob.model = m
	return ob
}

func (m *model) NewUid() orm.Uid {
	return orm.NewUid() // 采用默认生成
}

func (m *model) UidValid(u orm.Uid) (err error) {
	if len(u) > 0 {
		return
	} else {
		err = orm.ErrUidEmpty
	}
	return
}

func (m *model) EnsureIndex(indexMap orm.Index) (err error) {
	// default attr
	index := mgo.Index{
		Key:        []string{},
		Unique:     false,
		DropDups:   false, // important
		Background: true,
		Sparse:     true,
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
	// DropDups
	if i, ok := indexMap["DropDups"]; ok == true {
		if v, ok := i.(bool); ok {
			index.DropDups = v
		}
	}
	// DefaultLanguage
	if i, ok := indexMap["DefaultLanguage"]; ok == true {
		if v, ok := i.(string); ok {
			index.DefaultLanguage = v
		}
	}
	// LanguageOverride
	if i, ok := indexMap["LanguageOverride"]; ok == true {
		if v, ok := i.(string); ok {
			index.LanguageOverride = v
		}
	}
	// Sparse
	if i, ok := indexMap["Sparse"]; ok == true {
		if v, ok := i.(bool); ok {
			index.Sparse = v
		}
	}
	err = m.collection.EnsureIndex(index)
	return
}

func (m *model) EnsureIndexWithTag(st interface{}) (err error) {
	// index
	var (
		fieldInfoLis []*orm.FieldInfo
	)
	if fieldInfoLis, err = orm.StructModelInfo(st); err != nil {
		return
	}
	for _, f := range fieldInfoLis {
		if f.Index == false && f.Unique == false && f.IndexText == false {
			continue
		}
		if f.Index == true {
			if err = m.EnsureIndex(orm.Index{
				"Key":      f.IndexKeys,
				"Unique":   false,
				"DropDups": false, // mongodb
			}); err != nil {
				return
			}
		}
		if f.Unique == true {
			if err = m.EnsureIndex(orm.Index{
				"Key":      f.UniqueKeys,
				"Unique":   true,
				"DropDups": true, // mongodb:important
			}); err != nil {
				return
			}
		}
		if f.IndexText == true {
			if len(f.IndexKeys) == 1 {
				if err = m.EnsureIndex(orm.Index{
					"Key":              []string{"$text:" + f.IndexKeys[0]},
					"DefaultLanguage":  "en",
					"LanguageOverride": "en",
				}); err != nil {
					return
				}
			} else {
				err = orm.ErrIndexTextParam
				return
			}
		}
	}
	return
}

// 事务
func (m *model) Begin() (orm.Trans, error) {
	return nil, nil
}
func (m *model) BeginLevel(level string) (orm.Trans, error) {
	return nil, nil
}
func (m *model) Rollback(t orm.Trans) error {
	return nil
}
func (m *model) Commit(t orm.Trans) error {
	return nil
}
func (m *model) AutoTrans(t orm.Trans) error {
	return nil
}

// 兼容sql
func (m *model) Exec(query string, args ...interface{}) (result orm.Result, err error) {
	return
}
