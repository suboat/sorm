package songo

import (
	"encoding/json"
)

// 文档：https://github.com/suboat/songo

// Tag key
const (
	// TagSep mean value
	TagSep = '$'
	// TagValNe 不等于
	TagValNe = `$ne$`
	// TagValNo 不等于 alias
	TagValNo = `$no$`
	// TagValLt 小于
	TagValLt = `$lt$`
	// TagValLte 小于等于
	TagValLte = `$lte$`
	// TagValGt 大于
	TagValGt = `$gt$`
	// TagValGte 大于等于
	TagValGte = `$gte$`
	// TagValLike like
	TagValLike = `$like$`
	// TagValText 全文搜索
	TagValText = `$text$`
	// TagQueryKeyOr mean key or
	TagQueryKeyOr = `$or$`
	// TagQueryKeyAnd and
	TagQueryKeyAnd = `$and$`
	// TagQueryKeyIn in
	TagQueryKeyIn = `$in$`
	// TagUpdateInc mean update
	TagUpdateInc = `$inc$` // 数据库级别增
)

const (
	// SQLValEq sql 等于
	SQLValEq = `=`
	// SQLValNe  不等于
	SQLValNe = `!=`
	// SQLValLike like
	SQLValLike = `LIKE`
	// SQLValLt 小于
	SQLValLt = `<`
	// SQLValLte 小于等于
	SQLValLte = `<=`
	// SQLValGt 大于
	SQLValGt = `>`
	// SQLValGte 大于等于
	SQLValGte = `>=`
	// SQLValAnd 与
	SQLValAnd = `AND`
	// SQLValOr 或
	SQLValOr = `OR`
)

var (
	// ParseMapMax 目前允许5层解析
	ParseMapMax = 5
)

// ParseSQL 将songo格式解析为sql
func ParseSQL(m map[string]interface{}, prefix int) (sql string, vals []interface{}, err error) {
	if err = isSongoMapValid(m); err != nil {
		return
	}
	sql, vals, err = parserSQL(m, prefix, "$")
	return
}

// ParseMysql 将songo格式解析为mysql
func ParseMysql(m map[string]interface{}, prefix int) (sql string, vals []interface{}, err error) {
	if err = isSongoMapValid(m); err != nil {
		return
	}
	sql, vals, err = parserSQL(m, prefix, "?")
	return
}

// ParseMgo 将songo格式解析为mgo的M格式
func ParseMgo(m map[string]interface{}) (d map[string]interface{}, err error) {
	if err = isSongoMapValid(m); err != nil {
		return
	}
	// copy map:因为map可能会被修改，操作副本
	var (
		b     []byte
		mCopy = make(map[string]interface{})
	)
	if b, err = json.Marshal(m); err != nil {
		return
	}
	if err = json.Unmarshal(b, &mCopy); err != nil {
		return
	}
	// parse
	d, err = parseMgo(mCopy)
	return
}
