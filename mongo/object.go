package mongo

import (
	"git.yichui.net/open/orm"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

type objects struct {
	collection *mgo.Collection // mgo table
	query      *mgo.Query      // query handler
	skip       int             //
	limit      int             //
	total      int             // 搜索结果总数
	m          bson.M          // 缓存搜索条件
	count      int             // 缓存搜索结果
}

// query检查
func (o *objects) queryCheck() {
	if o.query == nil {
		o.query = o.collection.Find(nil)
	}
}

// 数目检查
func (o *objects) countCheck() {
	o.count, _ = o.query.Count()
}

// 保存至
func (o *objects) All(result interface{}) error {
	if o.query == nil {
		o.query = o.collection.Find(nil)
		if n, err := o.Count(); err == nil {
			o.total = n
		}
	}
	return o.query.All(result)
}

// 保存一条记录至
func (o *objects) One(result interface{}) (err error) {
	o.countCheck()
	if o.count >= 1 {
		err = o.query.One(result)
		if err == nil && o.count > 1 {
			err = orm.ErrFetchOneDuplicate // 找到多条记录
		}
	} else {
		err = orm.ErrNotExist
	}
	return err
}

// 总数
func (o *objects) Count() (int, error) {
	var err error
	if o.query == nil {
		o.count, err = o.collection.Count()
		o.total = o.count
		return o.count, err
	} else {
		o.count, err = o.query.Count()
		return o.count, err
	}
}

// escape % to regrex
func escapeRegrex(s string) interface{} {
	if len(s) > 2 && s[0] == '%' && s[len(s)-1] == '%' {
		return bson.M{"$regex": bson.RegEx{s[1 : len(s)-1], "i"}}
	}
	return s
}

// 搜索预处理
func filterM(val reflect.Value) (err error) {
	switch val.Kind() {
	case reflect.Interface:
		if val.CanSet() == false {
			err = orm.ErrReflectValSet
			return
		}
		// 处理模糊搜索
		if val.Elem().Kind() == reflect.String {
			// 过滤 % 为bson对象
			s := val.Elem().String()
			if len(s) > 2 && s[0] == '%' && s[len(s)-1] == '%' {
				newVal := bson.M{"$regex": bson.RegEx{s[1 : len(s)-1], "i"}}
				val.Set(reflect.ValueOf(newVal))
			}
		} else {
			// 复制出可编辑对象
			newVal := reflect.New(val.Elem().Type()).Elem()
			newVal.Set(val.Elem())
			filterM(newVal)
			val.Set(newVal)
		}
	case reflect.Map:
		if val.CanSet() == false {
			err = orm.ErrReflectValSet
			return
		}
		for _, k := range val.MapKeys() {
			newVal := reflect.New(val.MapIndex(k).Type()).Elem()
			newVal.Set(val.MapIndex(k))
			filterM(newVal)
			// 转义key, 如 %in% -> $in
			s := k.String()
			if len(s) > 2 && s[0] == '%' && s[len(s)-1] == '%' {
				sn := "$" + s[1:len(s)-1]
				val.SetMapIndex(k, reflect.Value{}) // delete key
				val.SetMapIndex(reflect.ValueOf(sn), newVal)
			} else {
				val.SetMapIndex(k, newVal)
			}
		}
	case reflect.Slice:
		if val.CanSet() == false {
			err = orm.ErrReflectValSet
			return
		}
		for i := 0; i < val.Len(); i += 1 {
			filterM(val.Index(i))
		}
	}
	// println("debug", val.Kind().String())
	return
}

// 搜索预处理
func preFilter(t orm.M) (ret bson.M) {
	ret = bson.M(t)
	// $ 字符处理(被angular过滤)
	for k, v := range ret {
		if k == "%or%" {
			ret["$or"] = v
			delete(ret, k)
		} else if k == "%and%" {
			ret["$and"] = v
			delete(ret, k)
		}
	}
	if err := filterM(reflect.ValueOf(&ret).Elem()); err != nil {
		panic(err)
	}
	return
}

// 搜索
func (o *objects) Filter(t orm.M) orm.Objects {
	o.m = preFilter(t)
	o.query = o.collection.Find(o.m)
	if n, err := o.Count(); err == nil {
		o.total = n
	}
	return o
}

// 排序
func (o *objects) Sort(fields ...string) orm.Objects {
	o.queryCheck()
	o.query = o.query.Sort(fields...)
	return o
}

// 摘要
func (o *objects) Meta() (*orm.Meta, error) {
	var (
		err error
	)
	_, err = o.Count() // 可能前面有limit操作，重新计数
	r := &orm.Meta{
		Limit:  o.limit,
		Skip:   o.skip,
		Total:  o.total,
		Length: o.count,
	}
	return r, err
}

// 限制
func (o *objects) Limit(n int) orm.Objects {
	o.queryCheck()
	o.query = o.query.Limit(n)
	o.limit = n
	return o
}

// 跳过
func (o *objects) Skip(n int) orm.Objects {
	if o.query == nil {
		o.query = o.collection.Find(nil)
	}
	o.query = o.query.Skip(n)
	o.skip = n
	return o
}

// 删除
func (o *objects) Delete() (err error) {
	if o.count == 0 {
		err = orm.ErrDeleteObjectEmpty
	} else if o.count == 1 {
		err = o.collection.Remove(o.m) // 删除一个记录
	} else {
		_, err = o.collection.RemoveAll(o.m) // 删除所有匹配记录
	}
	return
}

// 插入记录
func (o *objects) Create(i interface{}) (err error) {
	err = o.collection.Insert(i)
	return
}

// 更新记录
// TODO: 接收map类型参数
func (o *objects) Update(i interface{}) (err error) {
	if o.count == 0 {
		err = orm.ErrUpdateObjectEmpty
	} else if o.count == 1 {
		err = o.collection.Update(o.m, i) // 更新一个记录
	} else {
		// multi update only works with $ operators
		_, err = o.collection.UpdateAll(o.m, bson.M{"$set": i}) // 更新所有匹配记录
	}
	return
}

// 更新记录, 1条
// TODO: 接收map类型参数
func (o *objects) UpdateOne(i interface{}) (err error) {
	if o.count == 0 {
		err = orm.ErrUpdateObjectEmpty
	} else if o.count == 1 {
		err = o.collection.Update(o.m, i) // 更新一个记录
	} else {
		err = orm.ErrUpdateOneObjectMult
	}
	return
}

// 事务操作
func (o *objects) TDelete(t orm.Trans) (err error) {
	return o.Delete()
}
func (o *objects) TCreate(i interface{}, t orm.Trans) (err error) {
	return o.Create(i)
}
func (o *objects) TUpdate(i interface{}, t orm.Trans) (err error) {
	return o.Update(i)
}
func (o *objects) TUpdateOne(i interface{}, t orm.Trans) (err error) {
	return o.UpdateOne(i)
}
