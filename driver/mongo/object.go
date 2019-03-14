package mongo

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"

	//"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"reflect"
	"strings"
)

type Objects struct {
	Model *Model     // model
	query *mgo.Query // query handler
	m     bson.M

	// query and meta
	skip  int //
	limit int //
	count int // total num of query
	nums  int // fetch num of query

	// filter
	queryM orm.M    // store filter regular
	sorts  []string // sort

	// cache
	err error
	log orm.Logger
}

// query检查
func (o *Objects) queryCheck() {
	if o.query == nil {
		o.query = o.Model.Collection.Find(nil)
	}
}

// 数目检查
func (o *Objects) countCheck() {
	if o.err != nil {
		return
	}
	if o.count == -1 {
		if o.query == nil {
			o.count, o.err = o.Model.Collection.Count()
		} else {
			o.count, o.err = o.query.Count()
		}
	}
}

// 保存至
func (o *Objects) All(result interface{}) (err error) {
	if o.err != nil {
		return o.err
	}
	o.countCheck()

	if err = o.query.All(result); err == nil {
		// nums
		v := reflect.Indirect(reflect.ValueOf(result))
		if v.Kind() == reflect.Slice {
			o.nums = v.Len()
		}
		// debug
		o.log.Debug("[MONGO ALL] ", o.m)
	}
	return
}

// 保存一条记录至
func (o *Objects) One(result interface{}) (err error) {
	if o.err != nil {
		err = o.err
		return
	}
	o.countCheck()

	if o.count >= 1 {
		err = o.query.One(result)
		if err == nil && o.count > 1 {
			err = orm.ErrMatchMultiple // 找到多条记录
		}
	} else {
		err = orm.ErrMatchNone
	}
	return err
}

// 总数
func (o *Objects) Count() (n int, err error) {
	o.count = -1
	o.countCheck()
	return o.count, err
	//if o.err != nil {
	//	return -1, o.err
	//}
	//if o.query == nil {
	//	o.count, err = o.Model.Collection.Count()
	//	return o.count, err
	//} else {
	//	o.count, err = o.query.Count()
	//	return o.count, err
	//}
}

// 搜索
func (o *Objects) Filter(t orm.M) orm.Objects {
	o.queryM = t // cache
	if m, err := orm.HookParseMgo(t); err != nil {
		o.err = err // cache
		return o
	} else {
		o.m = bson.M(m)
	}
	o.query = o.Model.Collection.Find(o.m)
	o.countCheck()
	return o
}

// 排序 小写
func (o *Objects) Sort(fields ...string) orm.Objects {
	fields_ := []string{}
	for _, _s := range fields {
		fields_ = append(fields_, strings.ToLower(_s))
	}
	o.queryCheck()
	o.query = o.query.Sort(fields_...)
	o.sorts = fields
	return o
}

// 摘要
func (ob *Objects) Meta() (mt *orm.Meta, err error) {
	if ob.err != nil {
		err = ob.err
		return
	}
	ob.countCheck()
	//if _, err = ob.Count(); err != nil {
	//	return
	//}
	mt = &orm.Meta{
		Limit: ob.limit,
		Skip:  ob.skip,
		Count: ob.count,
		Num:   ob.nums,
	}
	// page
	if mt.Limit > 0 {
		mt.Page = mt.Skip / mt.Limit
	}
	// key
	if ob.queryM != nil {
		mt.Key = ob.queryM
	}
	// sort
	if ob.sorts != nil {
		mt.Sort = ob.sorts
	}
	return
}

// 限制
func (o *Objects) Limit(n int) orm.Objects {
	o.queryCheck()
	o.query = o.query.Limit(n)
	o.limit = n
	return o
}

// 跳过
func (o *Objects) Skip(n int) orm.Objects {
	o.queryCheck()
	o.query = o.query.Skip(n)
	o.skip = n
	return o
}

// 删除
func (o *Objects) Delete() (err error) {
	o.countCheck()
	if o.count == 0 {
		err = orm.ErrMatchNone
	} else if o.count == 1 {
		err = o.Model.Collection.Remove(o.m) // 删除一个记录
	} else {
		_, err = o.Model.Collection.RemoveAll(o.m) // 删除所有匹配记录
	}
	return
}

// 删除一条记录
func (ob *Objects) DeleteOne() (err error) {
	ob.countCheck()
	if ob.count == 0 {
		err = orm.ErrMatchNone
	} else if ob.count == 1 {
		err = ob.Delete()
	} else {
		err = orm.ErrMatchMultiple
	}
	return
}

// 插入记录
func (o *Objects) Create(i interface{}) (err error) {
	err = o.Model.Collection.Insert(i)
	return
}

// 更新记录
func (o *Objects) Update(i interface{}) (err error) {
	o.countCheck()
	if o.count == 0 {
		err = orm.ErrMatchNone
	} else if o.count == 1 {
		switch val := i.(type) {
		case map[string]interface{}:
			// map参数类型,更新特定值
			if val, err = orm.HookParseMgo(val); err != nil {
				return
			}
			// 特殊操作处理
			if _, ok := val["$set"]; ok == true {
				err = o.Model.Collection.Update(o.m, bson.M(val))
			} else if _, ok := val["$inc"]; ok == true {
				err = o.Model.Collection.Update(o.m, bson.M(val))
			} else {
				err = o.Model.Collection.Update(o.m, bson.M{"$set": val})
			}
		default:
			// 覆盖更新
			err = o.Model.Collection.Update(o.m, i)
		}
	} else {
		// multi update only works with $ operators
		_, err = o.Model.Collection.UpdateAll(o.m, bson.M{"$set": i}) // 更新所有匹配记录
	}
	return
}

// 更新记录, 1条
func (o *Objects) UpdateOne(i interface{}) (err error) {
	o.countCheck()
	if o.count == 0 {
		err = orm.ErrMatchNone
	} else if o.count == 1 {
		switch val := i.(type) {
		case map[string]interface{}:
			// map参数类型,更新特定值
			if val, err = orm.HookParseMgo(val); err != nil {
				return
			}
			// 特殊操作处理
			if _, ok := val["$set"]; ok == true {
				err = o.Model.Collection.Update(o.m, bson.M(val))
			} else if _, ok := val["$inc"]; ok == true {
				err = o.Model.Collection.Update(o.m, bson.M(val))
			} else {
				err = o.Model.Collection.Update(o.m, bson.M{"$set": val})
			}
		default:
			// 覆盖更新
			err = o.Model.Collection.Update(o.m, i)
		}
	} else {
		err = orm.ErrMatchMultiple
	}
	return
}

// 事务操作
func (o *Objects) TDelete(t orm.Trans) (err error) {
	if CfgTxUnsafe == false {
		err = orm.ErrTransNotSupport
	} else {
		err = o.Delete()
	}
	return
}
func (o *Objects) TCreate(i interface{}, t orm.Trans) (err error) {
	if CfgTxUnsafe == false {
		err = orm.ErrTransNotSupport
	} else {
		err = o.Create(i)
	}
	return
}
func (o *Objects) TUpdate(i interface{}, t orm.Trans) (err error) {
	if CfgTxUnsafe == false {
		err = orm.ErrTransNotSupport
	} else {
		err = o.Update(i)
	}
	return
}
func (o *Objects) TUpdateOne(i interface{}, t orm.Trans) (err error) {
	if CfgTxUnsafe == false {
		err = orm.ErrTransNotSupport
	} else {
		err = o.UpdateOne(i)
	}
	return
}
func (o *Objects) TAll(i interface{}, t orm.Trans) (err error) {
	if CfgTxUnsafe == false {
		err = orm.ErrTransNotSupport
	} else {
		err = o.All(i)
	}
	return
}
func (o *Objects) TOne(i interface{}, t orm.Trans) (err error) {
	if CfgTxUnsafe == false {
		err = orm.ErrTransNotSupport
	} else {
		err = o.One(i)
	}
	return
}
func (o *Objects) TCount(t orm.Trans) (n int, err error) {
	if CfgTxUnsafe == false {
		err = orm.ErrTransNotSupport
	} else {
		n, err = o.Count()
	}
	return
}
func (o *Objects) TDeleteOne(t orm.Trans) (err error) {
	if CfgTxUnsafe == false {
		err = orm.ErrTransNotSupport
	} else {
		err = o.DeleteOne()
	}
	return
}

// 兼容
func (o *Objects) TLockUpdate(t orm.Trans) (err error) {
	return
}

// sql 兼容
func (o *Objects) GetResult() (r orm.Result, err error) {
	return
}

// Copy 全拷贝
func (ob *Objects) Copy() (ret *Objects) {
	ret = new(Objects)
	*ret = *ob
	if ob.log != nil {
		if _log, ok := ob.log.(*log.Logger); ok {
			ret.log = _log.Copy()
		}
	}
	if ret.log == nil {
		if ret.log = orm.Log; ret.log == nil {
			ret.log = log.Log.Copy()
		}
	}
	return
}

// With 设置日志级别
func (ob *Objects) With(arg *orm.ArgObjects) (ret orm.Objects) {
	r := ob.Copy()
	if arg != nil {
		if arg.LogLevel > 0 {
			r.log.SetLevel(arg.LogLevel)
		}
	}
	return r
}
