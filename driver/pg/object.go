package pg

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"
	"github.com/suboat/sorm/songo"

	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Objects 实现
type Objects struct {
	Model  *Model
	Result orm.Result
	log    orm.Logger

	// query and meta
	skip  int //
	limit int //
	count int // total num of query
	nums  int // fetch num of query

	// filter
	queryM orm.M    // store filter regular
	sorts  []string // sort
	group  []string // group

	// cache: Query
	cacheQueryClean  bool          // if true, update cacheQuery* mandatorily next time
	cacheQueryExist  bool          // if true, query cache exist
	cacheQueryWhere  string        // sql contig after "where"
	cacheQueryValues []interface{} // cache query value list
	cacheQueryLimit  string        // sql contig from "limit"
	cacheQueryOrder  string        // sql contig from "order by"
	cacheQueryGroup  string        // sql contig from "group by"
}

//
type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

// GetResult 取结果
func (ob *Objects) GetResult() (res orm.Result, err error) {
	res = ob.Result
	return
}

// Filter 筛选
func (ob *Objects) Filter(t orm.M) orm.Objects {
	if t == nil {
		return ob
	}
	if ob.queryM == nil {
		ob.queryM = t
		return ob
	}
	_ = ob.queryM.Update(&t)
	ob.cacheQueryClean = true
	ob.count = -1
	return ob
}

// Limit 限制返回
func (ob *Objects) Limit(n int) orm.Objects {
	if n > -1 {
		ob.limit = n
		ob.cacheQueryLimit = fmt.Sprintf(`LIMIT %d OFFSET %d`, ob.limit, ob.skip)
	}
	return ob
}

// Skip 跳过记录
func (ob *Objects) Skip(n int) orm.Objects {
	if n > -1 {
		ob.skip = n
		ob.cacheQueryLimit = fmt.Sprintf(`LIMIT %d OFFSET %d`, ob.limit, ob.skip)
	}
	return ob
}

// Sort 用字段排序
func (ob *Objects) Sort(fields ...string) orm.Objects {
	for _, s := range fields {
		s = songo.SafeField(s)
		order := `ASC`
		if len(s) == 0 {
			continue
		}
		c := int(rune(s[0]))
		if c == '-' {
			order = `DESC`
			s = s[1:]
		} else if c == '+' {
			s = s[1:]
		}
		if len(s) == 0 {
			continue
		}
		if len(ob.cacheQueryOrder) == 0 {
			ob.cacheQueryOrder = fmt.Sprintf(`ORDER BY %s %s`, PubFieldWrap(s), order)
		} else {
			ob.cacheQueryOrder += fmt.Sprintf(`,%s %s`, PubFieldWrap(s), order)
		}
	}
	ob.sorts = append(ob.sorts, fields...)
	return ob
}

// 去重
func (ob *Objects) Group(fields ...string) orm.Objects {
	for _, s := range fields {
		s = songo.SafeField(s)
		if len(ob.cacheQueryGroup) == 0 {
			ob.cacheQueryGroup = fmt.Sprintf(`GROUP BY %s`, PubFieldWrap(s))
		} else {
			ob.cacheQueryGroup += fmt.Sprintf(`,%s`, PubFieldWrap(s))
		}
	}
	ob.group = append(ob.group, fields...)
	return ob
}

// Meta 信息
func (ob *Objects) Meta() (mt *orm.Meta, err error) {
	// may the 'limit' operating front, recount cache.
	if _, err = ob.Count(); err != nil {
		return
	}
	// upset nums
	if ob.nums == -1 {
		if ob.limit <= 0 {
			ob.nums = ob.count
		} else {
			if ob.count > ob.limit {
				if ob.nums = ob.count - ob.skip; ob.nums > ob.limit {
					ob.nums = ob.limit
				}
			} else if ob.nums > ob.count && ob.count > -1 {
				ob.nums = ob.count
			}
			if ob.nums == -1 {
				ob.nums = ob.count
			}
		}
	}

	// meta info
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
	// group
	if ob.group != nil {
		mt.Group = ob.group
	}
	return
}

// All fetch to
func (ob *Objects) All(result interface{}) (err error) {
	return ob.all(ob.Model.DatabaseSQL.DB, result)
}

// TAll 在事务中获取
func (ob *Objects) TAll(result interface{}, _t orm.Trans) (err error) {
	if _t == nil {
		return orm.ErrTransEmpty
	}
	t := _t.(*Trans)
	if t == nil {
		return orm.ErrTransInvalid
	}
	return ob.all(t, result)
}

// group-by https://stackoverflow.com/questions/1769361/postgresql-group-by-different-from-mysql
// https://stackoverflow.com/questions/17673457/converting-select-distinct-on-queries-from-postgresql-to-mysql
func (ob *Objects) all(ex execer, result interface{}) (err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}
	if ob.limit == 0 {
		ob.cacheQueryLimit = fmt.Sprintf(`OFFSET %d`, ob.skip)
	}
	//
	var (
		sqlCmd string
		fields = "*"
	)
	if ob.Model.DatabaseSQL.Unsafe == false {
		fields = strings.Join(PubFieldWrapByDest(result), ",")
	}
	if len(ob.group) > 0 {
		// group在postgres下的实现
		fields = fmt.Sprintf(`DISTINCT ON (%s) %s`, strings.Join(PubFieldWrapAll(ob.group), ","), fields)
	}
	if len(ob.cacheQueryWhere) == 0 {
		// select all
		sqlCmd = fmt.Sprintf(`SELECT %s FROM %s %s %s`,
			fields, ob.Model.GetTable(), ob.cacheQueryOrder, ob.cacheQueryLimit)
		sqlCmd = strings.ReplaceAll(sqlCmd, "  ", " ")
		if err = ex.Select(result, sqlCmd); err != nil {
			ob.log.Errorf(`[sql-all] %s err: %v`, sqlCmd, err)
		}
	} else {
		// select query
		sqlCmd = fmt.Sprintf(`SELECT %s FROM %s WHERE %s %s %s`,
			fields, ob.Model.GetTable(), ob.cacheQueryWhere, ob.cacheQueryOrder, ob.cacheQueryLimit)
		sqlCmd = strings.ReplaceAll(sqlCmd, "  ", " ")
		if err = ex.Select(result, sqlCmd, ob.cacheQueryValues...); err != nil {
			ob.log.Errorf(`[sql-all] %s VAL: %v err: %v`, sqlCmd, ob.cacheQueryValues, err)
		}
	}

	// count
	if err == nil {
		v := reflect.Indirect(reflect.ValueOf(result))
		if v.Kind() == reflect.Slice {
			ob.nums = v.Len()
		}
		// debug
		ob.log.Debugf(`[sql-all] %s`, sqlCmd)
	}

	return
}

// Count 统计
func (ob *Objects) Count() (num int, err error) {
	num, err = ob.countDo(ob.Model.DatabaseSQL.DB)
	return
}

// TCount is Count in transaction
func (ob *Objects) TCount(_t orm.Trans) (num int, err error) {
	if _t == nil {
		return 0, orm.ErrTransEmpty
	}
	t := _t.(*Trans)
	if t == nil {
		return 0, orm.ErrTransInvalid
	}
	return ob.countDo(t)
}

// Count
func (ob *Objects) countDo(ex execer) (num int, err error) {
	if ob.count > -1 {
		num = ob.count
		return
	}
	if err = ob.updateQuery(); err != nil {
		return
	}
	var (
		sqlCmd string
		fields = "*"
		where  = ""
	)
	if len(ob.cacheQueryWhere) > 0 {
		where = fmt.Sprintf("WHERE %s", ob.cacheQueryWhere)
	}
	if len(ob.group) > 0 {
		fields = fmt.Sprintf(`DISTINCT (%s)`, strings.Join(PubFieldWrapAll(ob.group), ","))
	}
	sqlCmd = fmt.Sprintf(`SELECT count(%s) FROM %s %s`, fields, ob.Model.GetTable(), where)
	sqlCmd = strings.ReplaceAll(sqlCmd, "  ", " ")
	if err = ex.Get(&num, sqlCmd, ob.cacheQueryValues...); err != nil {
		ob.log.Errorf(`[sql-count] %s VAL: %v err: %v`, sqlCmd, ob.cacheQueryValues, err)
	}

	if err == nil {
		ob.count = num
		// debug
		if num == 0 {
			ob.log.Debugf(`[sql-count-zero] %s %s`, ob.cacheQueryWhere, ob.cacheQueryValues)
		} else {
			// debug
			ob.log.Debugf(`[sql-count] %d: %s %s`, num, sqlCmd, ob.cacheQueryValues)
		}
	}
	return
}

// TOne fetch one to (in tx)
func (ob *Objects) TOne(result interface{}, _t orm.Trans) (err error) {
	if _, err = ob.TCount(_t); err != nil {
		return
	}
	if ob.count == 1 {
		var (
			sqlCmd string
			field  = "*"
		)
		if ob.Model.DatabaseSQL.Unsafe == false {
			field = strings.Join(PubFieldWrapByDest(result), ",")
		}
		//
		if _t == nil {
			return orm.ErrTransEmpty
		}
		t := _t.(*Trans)
		if t == nil {
			return orm.ErrTransInvalid
		}
		//
		sqlCmd = fmt.Sprintf(`SELECT %s FROM %s WHERE %s %s %s`,
			field, ob.Model.GetTable(), ob.cacheQueryWhere, ob.cacheQueryOrder, ob.cacheQueryLimit)
		err = t.Get(result, sqlCmd, ob.cacheQueryValues...)
		if err != nil {
			ob.log.Errorf(`[sql-one-t] %s VAL: %v err: %v`, sqlCmd, ob.cacheQueryValues, err)
		} else {
			ob.log.Debugf(`[sql-one-t] %s VAL: %v`, sqlCmd, ob.cacheQueryValues)
		}
	} else if ob.count == 0 {
		err = orm.ErrMatchNone
	} else {
		err = orm.ErrMatchMultiple
	}
	return
}

// One 取一条记录
// pg issue: missing destination name https://github.com/jmoiron/sqlx/issues/143
func (ob *Objects) One(result interface{}) (err error) {
	if _, err = ob.Count(); err != nil {
		return
	}
	if ob.count == 1 {
		var (
			sqlCmd string
			field  = "*"
		)
		if ob.Model.DatabaseSQL.Unsafe == false {
			field = strings.Join(PubFieldWrapByDest(result), ",")
		}
		sqlCmd = fmt.Sprintf(`SELECT %s FROM %s WHERE %s %s %s`,
			field, ob.Model.GetTable(), ob.cacheQueryWhere, ob.cacheQueryOrder, ob.cacheQueryLimit)
		sqlCmd = strings.ReplaceAll(sqlCmd, "  ", " ")
		err = ob.Model.DatabaseSQL.DB.Get(result, sqlCmd, ob.cacheQueryValues...)
		if err != nil {
			ob.log.Errorf(`[sql-one] %s VAL: %v err: %v`, sqlCmd, ob.cacheQueryValues, err)
		} else {
			ob.log.Debugf(`[sql-one] %s VAL: %v`, sqlCmd, ob.cacheQueryValues)
		}
	} else if ob.count == 0 {
		err = orm.ErrMatchNone
	} else {
		err = orm.ErrMatchMultiple
	}
	return
}

// Create 新建记录
func (ob *Objects) Create(insert interface{}) (err error) {
	return ob.create(ob.Model.DatabaseSQL.DB, insert)
}

//
func (ob *Objects) create(ex execer, insert interface{}) (err error) {
	if err = ob.Model.ContigParse(insert); err != nil {
		return
	}
	sqlCmd := fmt.Sprintf(`INSERT INTO %s;`, ob.Model.ContigInsert)
	ob.Result, err = ex.NamedExec(sqlCmd, insert)
	if err != nil {
		ob.log.Errorf(`[sql-create] %s VAL: %v err: %v`, sqlCmd, insert, err)
	} else {
		ob.log.Debugf(`[sql-create] %s VAL: %v`, sqlCmd, insert)
	}
	return
}

// Update 更新
func (ob *Objects) Update(record interface{}) (err error) {
	return ob.update(ob.Model.DatabaseSQL.DB, record)
}

//
func (ob *Objects) update(ex execer, record interface{}) (err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}

	reVal := reflect.Indirect(reflect.ValueOf(record))
	if reVal.Kind() == reflect.Map {
		// map: update set
		if m, ok := record.(map[string]interface{}); ok {
			// toLower
			_m := map[string]interface{}{}
			for _k, _v := range m {
				_m[strings.ToLower(_k)] = _v
			}
			m = _m
			_m = nil

			var (
				query      string
				queryWhere string
				args       []interface{}
				setLis     = []string{}
				fixMap     = map[string]interface{}{}
			)

			for k, v := range m {
				// simple prevent SQL injection
				if strings.Contains(k, " ") {
					err = orm.ErrUpdateMapKeyInvalid
					return
				}
				switch k {
				case orm.TagUpdateInc:
					if _m, _ok := v.(map[string]interface{}); _ok {
						for _k, _v := range _m {
							_k = strings.ToLower(_k)
							setLis = append(setLis, fmt.Sprintf(`"%s"="%s"+:%s`, _k, _k, _k))
							fixMap[_k] = _v
						}
					} else {
						err = orm.ErrUpdateIncValueInvalid
						return
					}
				default:
					setLis = append(setLis, fmt.Sprintf(`"%s"=:%s`, k, k))
				}
			}

			// fix
			delete(m, orm.TagUpdateInc)
			for k, v := range fixMap {
				m[k] = v
			}

			query = fmt.Sprintf(`UPDATE %s SET %s`, ob.Model.GetTable(), strings.Join(setLis, ", "))

			if len(ob.cacheQueryWhere) == 0 {
				// update one or more or nil
				if ob.Result, err = ex.NamedExec(query, m); err != nil {
					ob.log.Errorf(`[sql-update] %s VAL: %v err: %v`, query, m, err)
				} else {
					ob.log.Debugf(`[sql-update] %s VAL: %v`, query, m)
				}
			} else {
				// reformat sql, fieldName:key -> fieldName:$1
				if query, args, err = ob.Model.DatabaseSQL.DB.BindNamed(query, m); err != nil {
					ob.log.Error(err)
					return
				}
				// map assert as json, slice assert as json
				for i, v := range args {
					if m, ok := v.(map[string]interface{}); ok {
						if b, _err := json.Marshal(m); _err == nil {
							args[i] = b
						} else {
							ob.log.Warnf(`[sql-update] change map to json bytea error: %v`, _err)
							err = _err
							return
						}
					}
					if s, ok := v.([]interface{}); ok {
						if b, _err := json.Marshal(s); _err == nil {
							args[i] = b
						} else {
							ob.log.Warnf(`[sql-update] change slice to json bytea error: %v`, _err)
							err = _err
							return
						}
					}
				}

				// TODO: performance
				if queryWhere, _, err = ob.queryM.SQL(driverName, len(args)); err != nil {
					return
				}
				query = ob.Model.DatabaseSQL.DB.Rebind(query + ` WHERE ` + queryWhere)
				args = append(args, ob.cacheQueryValues...)
				if ob.Result, err = ex.Exec(query, args...); err != nil {
					ob.log.Errorf(`[sql-update] %s VAL: %v err: %v`, query, args, err)
				} else {
					ob.log.Debugf(`[sql-update] %s VAL: %v`, query, args)
				}
			}
		} else {
			err = orm.ErrUpdateMapTypeUnknown
			return
		}
	} else {
		// struct: overwrite
		sqlCmd := `UPDATE ` + ob.Model.ContigUpdate
		if len(ob.cacheQueryWhere) == 0 {
			ob.Result, err = ex.NamedExec(sqlCmd, record)
		} else {
			var (
				query      string
				queryWhere string
				args       []interface{}
			)
			if query, args, err = ob.Model.DatabaseSQL.DB.BindNamed(`UPDATE `+ob.Model.ContigUpdate, record); err != nil {
				return
			}
			// use special where contig
			// TODO: performance
			if queryWhere, _, err = ob.queryM.SQL(driverName, len(args)); err != nil {
				return
			}
			sqlCmd = ob.Model.DatabaseSQL.DB.Rebind(query + ` WHERE ` + queryWhere)
			args = append(args, ob.cacheQueryValues...)
			// exec
			if ob.Result, err = ex.Exec(sqlCmd, args...); err != nil {
				ob.log.Errorf(`[sql-update] %s VAL: %v err: %v`, sqlCmd, args, err)
			}
		}
		ob.log.Debugf(`[sql-update] %s`, sqlCmd)
	}

	return
}

// Delete 删除
func (ob *Objects) Delete() error {
	return ob.delete(ob.Model.DatabaseSQL.DB)
}

// DeleteOne 删除一条记录
func (ob *Objects) DeleteOne() (err error) {
	if _, err = ob.Count(); err != nil {
		return
	}
	if ob.count == 0 {
		err = orm.ErrMatchNone
	} else if ob.count == 1 {
		err = ob.delete(ob.Model.DatabaseSQL.DB)
	} else {
		err = orm.ErrMatchMultiple
	}
	return
}

func (ob *Objects) delete(ex execer) (err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}
	sqlCmd := ``
	if len(ob.cacheQueryWhere) == 0 {
		// delete all record
		sqlCmd = fmt.Sprintf(`DELETE FROM %s`, ob.Model.GetTable())
		if ob.Result, err = ex.Exec(sqlCmd); err != nil {
			ob.log.Errorf(`[sql-delete] %s err: %v`, sqlCmd, err)
		}
	} else {
		sqlCmd = fmt.Sprintf(`DELETE FROM %s WHERE %s`, ob.Model.GetTable(), ob.cacheQueryWhere)
		if ob.Result, err = ex.Exec(sqlCmd, ob.cacheQueryValues...); err != nil {
			ob.log.Errorf(`[sql-delete] %s VAL: %v err: %v`, sqlCmd, ob.cacheQueryValues, err)
		}
	}
	ob.log.Debugf(`[sql-delete] %s`, sqlCmd)
	return
}

// UpdateOne 更新一条记录
func (ob *Objects) UpdateOne(record interface{}) (err error) {
	if _, err = ob.Count(); err != nil {
		return
	}
	if ob.count == 0 {
		err = orm.ErrMatchNone
	} else if ob.count == 1 {
		err = ob.Update(record)
	} else {
		err = orm.ErrMatchMultiple
	}
	return
}

// TDelete 事务中删除
func (ob *Objects) TDelete(_t orm.Trans) (err error) {
	if _t == nil {
		return orm.ErrTransEmpty
	}
	t := _t.(*Trans)
	if t == nil {
		return orm.ErrTransInvalid
	}
	err = ob.delete(t)
	_ = t.DebugPush(`[delete]` + ob.cacheQueryWhere)
	return
}

// TDeleteOne 事务中删除
func (ob *Objects) TDeleteOne(_t orm.Trans) (err error) {
	if _, err = ob.Count(); err != nil {
		return
	}
	if _t == nil {
		return orm.ErrTransEmpty
	}
	t := _t.(*Trans)
	if t == nil {
		return orm.ErrTransInvalid
	}
	if ob.count == 0 {
		err = orm.ErrMatchNone
	} else if ob.count == 1 {
		err = ob.delete(t)
	} else {
		err = orm.ErrMatchMultiple
	}
	_ = t.DebugPush(`[deleteOne]` + ob.cacheQueryWhere)
	return
}

// TCreate 事务中创建
func (ob *Objects) TCreate(insert interface{}, _t orm.Trans) (err error) {
	if _t == nil {
		return orm.ErrTransEmpty
	}
	t := _t.(*Trans)
	if t == nil {
		return orm.ErrTransInvalid
	}
	err = ob.create(t, insert)
	_ = t.DebugPush(`[create]` + ob.cacheQueryWhere)
	return
}

// TUpdate 事务中更新
func (ob *Objects) TUpdate(record interface{}, _t orm.Trans) (err error) {
	if _t == nil {
		return orm.ErrTransEmpty
	}
	t := _t.(*Trans)
	if t == nil {
		return orm.ErrTransInvalid
	}
	err = ob.update(t, record)
	_ = t.DebugPush(`[update]` + ob.cacheQueryWhere)
	return
}

// TUpdateOne 事务中更新
func (ob *Objects) TUpdateOne(record interface{}, _t orm.Trans) (err error) {
	if _, err = ob.Count(); err != nil {
		return
	}
	if _t == nil {
		return orm.ErrTransEmpty
	}
	t := _t.(*Trans)
	if ob.count == 0 {
		err = orm.ErrMatchNone
	} else if ob.count == 1 {
		err = ob.TUpdate(record, t)
	} else {
		err = orm.ErrMatchMultiple
	}
	_ = t.DebugPush(`[updateOne]` + ob.cacheQueryWhere)
	return
}

// build in method
// update query cache
func (ob *Objects) updateQuery() (err error) {
	//log.Warn("ob.cacheQueryValues ", ob.cacheQueryWhere, " ", ob.cacheQueryValues)
	// query all
	if ob.queryM == nil {
		ob.cacheQueryValues = nil
		ob.cacheQueryWhere = ""
		return
	}
	// use cache
	if ob.cacheQueryExist && !ob.cacheQueryClean {
		return
	}
	// update
	if ob.cacheQueryWhere, ob.cacheQueryValues, err = ob.queryM.SQL(driverName, 0); err != nil {
		return
	}
	//log.Warn("ob.cacheQueryValues ", ob.cacheQueryWhere, " ", ob.cacheQueryValues)
	ob.cacheQueryExist = true
	ob.cacheQueryClean = false
	return
}

// TLockUpdate row lock
func (ob *Objects) TLockUpdate(t orm.Trans) (err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}

	var sqlCmd string
	if len(ob.cacheQueryWhere) == 0 {
		// select all
		err = orm.ErrTransLockWholeTable
		return
	}
	sqlCmd = fmt.Sprintf(`SELECT * FROM %s WHERE %s FOR UPDATE`, ob.Model.GetTable(), ob.cacheQueryWhere)

	if ob.Result, err = t.Exec(sqlCmd, ob.cacheQueryValues...); err != nil {
		ob.log.Errorf(`[sql-lock-t] %s VAL: %v err: %v`, sqlCmd, ob.cacheQueryValues, err)
	} else {
		ob.log.Debugf(`[sql-lock-t] %s`, sqlCmd)
	}
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
