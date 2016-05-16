package mongo

import (
	"git.yichui.net/open/orm"

	"gopkg.in/mgo.v2"
	"reflect"
	"strings"
	"sync"
)

type model struct {
	name string
	sync.RWMutex
	collection *mgo.Collection
	objects    *objects
}

func (m *model) String() string {
	return m.name
}

func (m *model) Objects() orm.Objects {
	if m.objects == nil {
		o := new(objects)
		o.collection = m.collection
		return o
	} else {
		return m.objects
	}
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

func parseIndexTag(st interface{}, m *map[string][]string) (err error) {
	stVal := reflect.Indirect(reflect.ValueOf(st))
	switch stVal.Kind() {
	case reflect.Struct:
		stType := stVal.Type()
		for i := 0; i < stType.NumField(); i++ {
			fType := stType.Field(i)
			fVal := stVal.Field(i)
			//println("debug", fType.Name, fVal.Kind().String(), fType.Tag.Get(orm.TagKey))
			for _, v := range strings.Split(fType.Tag.Get(orm.TagKey), ";") {
				if len(v) > 0 {
					n := strings.ToLower(fType.Name)
					if _, ok := (*m)[v]; ok == false {
						(*m)[v] = []string{}
					}
					(*m)[v] = append((*m)[v], n)
				}
			}
			if fVal.Kind() == reflect.Struct {
				parseIndexTag(fVal.Interface(), m)
			}
		}
		break
	default:
		break
	}
	return
}

func (m *model) EnsureIndexWithTag(st interface{}) (err error) {
	k := make(map[string][]string)
	if err = parseIndexTag(st, &k); err != nil {
		panic(err)
	}
	for i, f := range k {
		switch i {
		case "index":
			for _, n := range f {
				if err = m.EnsureIndex(orm.Index{
					"Key":      []string{n},
					"Unique":   false,
					"DropDups": false,
				}); err != nil {
					return
				}
			}
			break
		case "unique":
			for _, n := range f {
				if err = m.EnsureIndex(orm.Index{
					"Key":      []string{n},
					"Unique":   true,
					"DropDups": true, // important
				}); err != nil {
					return
				}
			}
			break
		case "text":
			for _, n := range f {
				if err = m.EnsureIndex(orm.Index{
					"Key":              []string{"$text:" + n},
					"DefaultLanguage":  "en",
					"LanguageOverride": "en",
				}); err != nil {
					return
				}
			}
			break
		default:
			//println("not suport index method:", i)
		}
	}
	return
}

// 事务
func (m *model) Begin() (orm.Trans, error) {
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
