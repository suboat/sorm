package postgres

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"

	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Objects struct {
	Model  *Model
	Result sql.Result

	// query and meta
	skip  int //
	limit int //
	total int // total num of query
	count int // fetch num of query

	// filter
	queryM orm.M // store filter regular

	// cache
	// cacheQuery
	cacheQueryClean  bool          // if true, update cacheQuery* mandatorily next time
	cacheQueryExist  bool          // if true, query cache exist
	cacheQueryWhere  string        // sql contig after "where"
	cacheQueryValues []interface{} // cache query valuse lis
	cacheQueryLimit  string        // sql contig from "limit"
	cacheQueryOrder  string        // sql contig from "order by"
}

type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
}

// Filter
func (ob *Objects) Filter(t orm.M) orm.Objects {
	if t == nil {
		return ob
	}
	if ob.queryM == nil {
		ob.queryM = t
		return ob
	}
	ob.queryM.Update(&t)
	ob.cacheQueryClean = true
	return ob
}

// Count
func (ob *Objects) Count() (num int, err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}
	var sqlCmd string
	if len(ob.cacheQueryWhere) == 0 {
		// select all
		sqlCmd = "SELECT count(*) FROM " + ob.Model.TableName
		err = ob.Model.DatabaseSql.DB.Get(&num, sqlCmd)
	} else {
		//sqlCmd = "SELECT count(*) FROM " + ob.Model.TableName + " WHERE " + ob.cacheQueryWhere + " " + ob.cacheQueryLimit
		sqlCmd = "SELECT count(*) FROM " + ob.Model.TableName + " WHERE " + ob.cacheQueryWhere // count have not limit
		err = ob.Model.DatabaseSql.DB.Get(&num, sqlCmd, ob.cacheQueryValues...)
	}
	if err == nil {
		ob.total = num
		// debug
		if num == 0 {
			log.Warn("count zero: ", ob.cacheQueryWhere, ob.cacheQueryValues)
		} else {
			// debug
			// "SELECT * FROM "+ob.Model.TableName+" WHERE "+ob.cacheQueryWhere+ob.cacheQueryOrder+" "+ob.cacheQueryLimit,
			//log.Warn("check: ", sqlCmd, " ", ob.cacheQueryValues)
		}
	}
	return
}

// Limit
func (ob *Objects) Limit(n int) orm.Objects {
	if n > -1 {
		ob.limit = n
		ob.cacheQueryLimit = fmt.Sprintf("LIMIT %d OFFSET %d", ob.limit, ob.skip)
	}
	return ob
}

// Skip
func (ob *Objects) Skip(n int) orm.Objects {
	if n > -1 {
		ob.skip = n
		ob.cacheQueryLimit = fmt.Sprintf("LIMIT %d OFFSET %d", ob.limit, ob.skip)
	}
	return ob
}

// Sort
func (ob *Objects) Sort(fields ...string) orm.Objects {
	for _, s := range fields {
		order := "ASC"
		if len(s) == 0 {
			continue
		}
		c := int(rune(s[0]))
		if c == '-' {
			order = "DESC"
			s = s[1:]
		} else if c == '+' {
			s = s[1:]
		}
		if len(s) == 0 {
			continue
		}
		if len(ob.cacheQueryOrder) == 0 {
			ob.cacheQueryOrder = fmt.Sprintf("ORDER BY %s %s", s, order)
		} else {
			ob.cacheQueryOrder += fmt.Sprintf(", %s %s", s, order)
		}
	}
	return ob
}

// Meta
func (ob *Objects) Meta() (mt *orm.Meta, err error) {
	// may the 'limit' operating front, recount cache.
	_, err = ob.Count()
	// meta info
	mt = &orm.Meta{
		Limit:  ob.limit,
		Skip:   ob.skip,
		Total:  ob.total,
		Length: ob.count,
	}
	return
}

// Fetch to
func (ob *Objects) All(result interface{}) (err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}
	if ob.limit == 0 {
		ob.cacheQueryLimit = fmt.Sprintf("OFFSET %d", ob.skip)
	}
	return ob.all(result)
}
func (ob *Objects) AllDebug(result interface{}) (err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}
	// debug
	if ob.limit == 0 || ob.limit > 1000 {
		if ob.limit > 1000 {
			log.Warn("!!!Debug limit set to 1000: ", " ", ob.limit, ob.cacheQueryLimit)
		}
		ob.cacheQueryLimit = fmt.Sprintf("LIMIT %d OFFSET %d", 1000, ob.skip)
	}
	return ob.all(result)
}
func (ob *Objects) all(result interface{}) (err error) {
	var _sql string
	if len(ob.cacheQueryWhere) == 0 {
		// select all
		_sql = "SELECT * FROM " + ob.Model.TableName + " " + ob.cacheQueryOrder + " " + ob.cacheQueryLimit
		err = ob.Model.DatabaseSql.DB.Select(result, _sql)
	} else {
		// select query
		_sql = "SELECT * FROM " + ob.Model.TableName + " WHERE " + ob.cacheQueryWhere + " " + ob.cacheQueryOrder + " " + ob.cacheQueryLimit
		err = ob.Model.DatabaseSql.DB.Select(result, _sql, ob.cacheQueryValues...)
	}

	// count
	if err == nil {
		v := reflect.Indirect(reflect.ValueOf(result))
		if v.Kind() == reflect.Slice {
			ob.count = v.Len()
		}
	}

	log.Debug("sql: ", _sql)
	return
}

// Fetch one to
// pg issue: missing destination name https://github.com/jmoiron/sqlx/issues/143
func (ob *Objects) One(result interface{}) (err error) {
	if _, err = ob.Count(); err != nil {
		return
	}
	if ob.total == 1 {
		//err = ob.Model.DatabaseSql.DB.Get(result,
		err = ob.Model.DatabaseSql.DB.Unsafe().Get(result,
			"SELECT * FROM "+ob.Model.TableName+" WHERE "+ob.cacheQueryWhere+ob.cacheQueryOrder+" "+ob.cacheQueryLimit,
			ob.cacheQueryValues...)
	} else if ob.total == 0 {
		err = orm.ErrNotExist
	} else {
		err = orm.ErrFetchOneDuplicate
		//log.Error("count: ", ob.total, " sql: ", ob.cacheQueryWhere+ob.cacheQueryOrder+" "+ob.cacheQueryLimit)
	}
	return
}

// Create
func (ob *Objects) Create(insert interface{}) (err error) {
	return ob.create(ob.Model.DatabaseSql.DB, insert)
}
func (ob *Objects) create(ex execer, insert interface{}) (err error) {
	if err = ob.Model.ContigParse(insert); err != nil {
		return
	}
	sqlCmd := fmt.Sprintf("INSERT INTO %s;", ob.Model.ContigInsert)
	ob.Result, err = ex.NamedExec(sqlCmd, insert)
	if err != nil {
		log.Error("sql: ", sqlCmd, " vals: ", insert)
	} else {
		log.Debug("sql: ", sqlCmd, " vals: ", insert)
	}
	return
}

// Update
func (ob *Objects) Update(record interface{}) (err error) {
	return ob.update(ob.Model.DatabaseSql.DB, record)
}
func (ob *Objects) update(ex execer, record interface{}) (err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}

	reVal := reflect.Indirect(reflect.ValueOf(record))
	if reVal.Kind() == reflect.Map {
		// map: update set
		if m, ok := record.(map[string]interface{}); ok {
			var (
				query      string
				queryWhere string
				args       []interface{}
				setLis     = []string{}
			)

			for k, _ := range m {
				// simple prevent SQL injection
				if strings.Index(k, " ") > -1 {
					err = orm.ErrMInvalid
					return
				}
				setLis = append(setLis, fmt.Sprintf("%s=:%s", k, k))
			}

			query = fmt.Sprintf("UPDATE %s SET %s", ob.Model.TableName, strings.Join(setLis, ", "))

			if len(ob.cacheQueryWhere) == 0 {
				// update one or more or nil
				ob.Result, err = ex.NamedExec(query, m)
			} else {
				// reformat sql, fieldName:key -> fieldName:$1
				if query, args, err = ob.Model.DatabaseSql.DB.BindNamed(query, m); err != nil {
					return
				}
				// map assert as json, slice assert as json
				for i, v := range args {
					if m, ok := v.(map[string]interface{}); ok {
						if b, _err := json.Marshal(m); err != nil {
							log.Debug("upadte: change map to json bytea error: ", _err.Error())
							err = _err
							return
						} else {
							args[i] = b
						}
					}
					if s, ok := v.([]interface{}); ok {
						if b, _err := json.Marshal(s); err != nil {
							log.Debug("upadte: change slice to json bytea error: ", _err.Error())
							err = _err
							return
						} else {
							args[i] = b
						}
					}
				}

				// TODO: performace
				if queryWhere, _, err = ob.queryM.Sql(DatabaseHash, len(args)); err != nil {
					return
				}
				query = ob.Model.DatabaseSql.DB.Rebind(query + " WHERE " + queryWhere)
				args = append(args, ob.cacheQueryValues...)
				log.Debug("update with map: ", query, " vals: ", args, " org: ", ob.cacheQueryValues)
				ob.Result, err = ex.Exec(query, args...)
			}

			return
		} else {
			err = ErrUpdateMapTyep
			return
		}
	} else {
		// struct: overwrite
		if len(ob.cacheQueryWhere) == 0 {
			ob.Result, err = ex.NamedExec("UPDATE "+ob.Model.ContigUpdate, record)
		} else {
			var (
				query      string
				queryWhere string
				args       []interface{}
			)
			if query, args, err = ob.Model.DatabaseSql.DB.BindNamed("UPDATE "+ob.Model.ContigUpdate, record); err != nil {
				return
			}
			// use special where contig
			// TODO: performace
			if queryWhere, _, err = ob.queryM.Sql(DatabaseHash, len(args)); err != nil {
				return
			}
			query = ob.Model.DatabaseSql.DB.Rebind(query + " WHERE " + queryWhere)
			args = append(args, ob.cacheQueryValues...)
			// exec
			//println("debug query", query)
			ob.Result, err = ex.Exec(query, args...)
		}
		return
	}

	return
}

// Delete
func (ob *Objects) Delete() error {
	return ob.delete(ob.Model.DatabaseSql.DB)
}
func (ob *Objects) delete(ex execer) (err error) {
	if err = ob.updateQuery(); err != nil {
		return
	}
	if len(ob.cacheQueryWhere) == 0 {
		// delete all record
		ob.Result, err = ex.Exec("DELETE FROM " + ob.Model.TableName)
		return
	} else {
		ob.Result, err = ex.Exec("DELETE FROM "+ob.Model.TableName+" WHERE "+ob.cacheQueryWhere, ob.cacheQueryValues...)
	}
	return
}

// other method

// Update
func (ob *Objects) UpdateOne(record interface{}) (err error) {
	if _, err = ob.Count(); err != nil {
		return
	}
	if ob.total == 0 {
		err = orm.ErrUpdateObjectEmpty
	} else if ob.total == 1 {
		err = ob.Update(record)
	} else {
		err = orm.ErrUpdateOneObjectMult
	}
	return
}

// transaction
func (ob *Objects) TDelete(t orm.Trans) (err error) {
	err = ob.delete(t)
	return
}
func (ob *Objects) TCreate(insert interface{}, t orm.Trans) (err error) {
	err = ob.create(t, insert)
	return
}
func (ob *Objects) TUpdate(record interface{}, t orm.Trans) (err error) {
	err = ob.update(t, record)
	return
}
func (ob *Objects) TUpdateOne(record interface{}, t orm.Trans) (err error) {
	if _, err = ob.Count(); err != nil {
		return
	}
	if ob.total == 0 {
		err = orm.ErrUpdateObjectEmpty
	} else if ob.total == 1 {
		err = ob.TUpdate(record, t)
	} else {
		err = orm.ErrUpdateOneObjectMult
	}
	return
}

// buildin method
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
	if ob.cacheQueryExist == true && ob.cacheQueryClean == false {
		return
	}
	// update
	if ob.cacheQueryWhere, ob.cacheQueryValues, err = ob.queryM.Sql(DatabaseHash); err != nil {
		return
	}
	//log.Warn("ob.cacheQueryValues ", ob.cacheQueryWhere, " ", ob.cacheQueryValues)
	ob.cacheQueryExist = true
	ob.cacheQueryClean = false
	return
}
