package songo

import (
	"github.com/suboat/sorm"

	"encoding/json"
	"fmt"
	"testing"
)

// https://github.com/suboat/songo 例子
var (
	// IsLog 为true时打印详细信息
	IsLog = false
	// 实例1
	songTestASql = `("age" >= $1) AND ("limit" = $2) AND ("skip" = $3)`
	songTestA1   = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		"age":   "$gte$30",
	}
	songTestA2 = map[string]interface{}{
		"limit":    10,
		"skip":     20,
		"age$gte$": 30,
	}
	songTestA3 = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		"age": map[string]interface{}{
			"$gte$": 30,
		},
	}
	// 实例2
	songTestBSql = `("age" >= $1 AND "age" <= $2) AND ("limit" = $3) AND ("skip" = $4)`
	songTestB1   = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		orm.TagQueryKeyAnd + "age": []interface{}{
			"$gte$30",
			"$lte$40",
		},
	}
	songTestB2 = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		orm.TagQueryKeyAnd: []interface{}{
			map[string]interface{}{
				"age": "$gte$30",
			},
			map[string]interface{}{
				"age": "$lte$40",
			},
		},
	}
	songTestB3 = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		orm.TagQueryKeyAnd: []interface{}{
			map[string]interface{}{
				"age$gte$": 30,
			},
			map[string]interface{}{
				"age$lte$": 40,
			},
		},
	}
	songTestB4 = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		orm.TagQueryKeyAnd: []interface{}{
			map[string]interface{}{
				"age": map[string]interface{}{
					"$gte$": 30,
				},
			},
			map[string]interface{}{
				"age": map[string]interface{}{
					"$lte$": 40,
				},
			},
		},
	}
	// 实例3
	songTestCSql = `("age" <= $1 OR "age" >= $2) AND ("limit" = $3) AND ("skip" = $4)`
	songTestC1   = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		orm.TagQueryKeyOr + "age": []interface{}{
			"$lte$30",
			"$gte$40",
		},
	}
	songTestC2 = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		orm.TagQueryKeyOr: []interface{}{
			map[string]interface{}{
				"age": "$lte$30",
			},
			map[string]interface{}{
				"age": "$gte$40",
			},
		},
	}
	songTestC3 = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		orm.TagQueryKeyOr: []interface{}{
			map[string]interface{}{
				"age$lte$": 30,
			},
			map[string]interface{}{
				"age$gte$": 40,
			},
		},
	}
	songTestC4 = map[string]interface{}{
		"limit": 10,
		"skip":  20,
		orm.TagQueryKeyOr: []interface{}{
			map[string]interface{}{
				"age": map[string]interface{}{
					"$lte$": 30,
				},
			},
			map[string]interface{}{
				"age": map[string]interface{}{
					"$gte$": 40,
				},
			},
		},
	}
	// 实例4.1
	songTestD1 = map[string]interface{}{
		"uid": "11111111-1111-1111-1111-111111111111",
		orm.TagQueryKeyOr: []interface{}{
			map[string]interface{}{
				orm.TagQueryKeyAnd: []interface{}{
					map[string]interface{}{
						"status": 0,
					},
					map[string]interface{}{
						"active": true,
					},
				},
			},
			map[string]interface{}{
				orm.TagQueryKeyAnd: []interface{}{
					map[string]interface{}{
						"status": 1,
						"active": false,
					},
				},
			},
		},
		orm.TagQueryKeyAnd: []interface{}{
			map[string]interface{}{
				orm.TagQueryKeyOr: []interface{}{
					map[string]interface{}{
						"category": "type1",
					},
					map[string]interface{}{
						"category": "type2",
					},
				},
			},
			map[string]interface{}{
				"freeze": false,
			},
		},
	}
	// 实例4.2
	songTestD2 = map[string]interface{}{
		"id": "001",
		orm.TagQueryKeyOr: []interface{}{
			map[string]interface{}{
				orm.TagQueryKeyAnd: []interface{}{
					map[string]interface{}{
						"freeze": true,
					},
					map[string]interface{}{
						"category": "type1",
						orm.TagQueryKeyOr: []interface{}{
							map[string]interface{}{
								"status": 1,
							},
							map[string]interface{}{
								"status": 2,
							},
						},
					},
				},
			},
			map[string]interface{}{
				orm.TagQueryKeyAnd: []interface{}{
					map[string]interface{}{
						"freeze":                     false,
						"category":                   "type2",
						orm.TagQueryKeyOr + "status": []interface{}{"1", 2},
					},
				},
			},
		},
	}
	// 实例4.3
	songTestD3 = map[string]interface{}{
		orm.TagQueryKeyOr + "category": []interface{}{"type1", "type2"},
		orm.TagQueryKeyOr + "status":   []interface{}{"1", 2},
	}
	// 验证
	songTestDSql = []string{
		// 实例4.1
		`((("category" = $1) OR ("category" = $2)) AND "freeze" = $3) AND ((("status" = $4) AND ("active" = $5)) OR (("active" = $6 AND "status" = $7))) AND ("uid" = $8)`,
		// 实例4.2
		`((("freeze" = $1) AND ((("status" = $2) OR ("status" = $3)) AND "category" = $4)) OR (((("status" = $5) OR ("status" = $6)) AND "category" = $7 AND "freeze" = $8))) AND ("id" = $9)`,
		//  实例4.3
		`("category" = $1 OR "category" = $2) AND ("status" = $3 OR "status" = $4)`,
	}
	songTestAMgo = []string{
		`{"age":{"$gte":"30"},"limit":10,"skip":20}`,
		`{"age":{"$gte":30},"limit":10,"skip":20}`,
		`{"age":{"$gte":30},"limit":10,"skip":20}`,
	}
	songTestBMgo = []string{
		`{"$and":[{"age":{"$gte":"30"}},{"age":{"$lte":"40"}}],"limit":10,"skip":20}`,
		`{"$and":[{"age":{"$gte":"30"}},{"age":{"$lte":"40"}}],"limit":10,"skip":20}`,
		`{"$and":[{"age":{"$gte":30}},{"age":{"$lte":40}}],"limit":10,"skip":20}`,
		`{"$and":[{"age":{"$gte":30}},{"age":{"$lte":40}}],"limit":10,"skip":20}`,
	}
	songTestCMgo = []string{
		`{"$or":[{"age":{"$lte":"30"}},{"age":{"$gte":"40"}}],"limit":10,"skip":20}`,
		`{"$or":[{"age":{"$lte":"30"}},{"age":{"$gte":"40"}}],"limit":10,"skip":20}`,
		`{"$or":[{"age":{"$lte":30}},{"age":{"$gte":40}}],"limit":10,"skip":20}`,
		`{"$or":[{"age":{"$lte":30}},{"age":{"$gte":40}}],"limit":10,"skip":20}`,
	}
	songTestDMgo = []string{
		// 实例4.1
		`{"$and":[{"$or":[{"category":"type1"},{"category":"type2"}]},{"freeze":false}],"$or":[{"$and":[{"status":0},{"active":true}]},{"$and":[{"active":false,"status":1}]}],"uid":"11111111-1111-1111-1111-111111111111"}`,
		// 实例4.2
		`{"$or":[{"$and":[{"freeze":true},{"$or":[{"status":1},{"status":2}],"category":"type1"}]},{"$and":[{"$or":[{"status":"1"},{"status":2}],"category":"type2","freeze":false}]}],"id":"001"}`,
		// 实例4.3
		`{"$and":[{"$or":[{"category":"type1"},{"category":"type2"}]},{"$or":[{"status":"1"},{"status":2}]}]}`,
	}
)

// 将songo格式解析为sql
func Test_SongoParseSql(t *testing.T) {
	// 实例1
	for i, m := range []map[string]interface{}{songTestA1, songTestA2, songTestA3} {
		if sql, vals, err := ParseSQL(m, 0); err != nil {
			t.Fatal(err)
		} else {
			b, _ := json.Marshal(m)
			v, _ := json.Marshal(vals)
			if sql != songTestASql {
				t.Logf("%s", songTestASql)
				t.Logf("%s %s", sql, string(v))
				t.Fatalf("A%d %s <- %s <- %s", i+1, sql, string(v), string(b))
			} else if IsLog {
				t.Logf("A%d %s <- %s <- %s", i+1, sql, string(v), string(b))
			}
		}
	}
	// 实例2
	for i, m := range []map[string]interface{}{songTestB1, songTestB2, songTestB3, songTestB4} {
		if sql, vals, err := ParseSQL(m, 0); err != nil {
			t.Fatal(err)
		} else {
			b, _ := json.Marshal(m)
			v, _ := json.Marshal(vals)
			if sql != songTestBSql {
				t.Logf("%s", songTestBSql)
				t.Logf("%s %s", sql, string(v))
				t.Fatalf("B%d %s <- %s <- %s", i+1, sql, string(v), string(b))
			} else if IsLog {
				t.Logf("B%d %s <- %s <- %s", i+1, sql, string(v), string(b))
			}
		}
	}
	// 实例3
	for i, m := range []map[string]interface{}{songTestC1, songTestC2, songTestC3, songTestC4} {
		if sql, vals, err := ParseSQL(m, 0); err != nil {
			t.Fatal(err)
		} else {
			b, _ := json.Marshal(m)
			v, _ := json.Marshal(vals)
			if sql != songTestCSql {
				t.Logf("%s", songTestCSql)
				t.Logf("%s %s", sql, string(v))
				t.Fatalf("C%d %s <- %s <- %s", i+1, sql, string(v), string(b))
			} else if IsLog {
				t.Logf("C%d %s <- %s <- %s", i+1, sql, string(v), string(b))
			}
		}
	}
	// 实例4
	for i, m := range []map[string]interface{}{songTestD1, songTestD2, songTestD3} {
		b, _ := json.Marshal(m)
		if sql, vals, err := ParseSQL(m, 0); err != nil {
			t.Fatalf("D%d %s %v", i+1, string(b), err)
		} else {
			v, _ := json.Marshal(vals)
			if sql != songTestDSql[i] {
				t.Logf("%s", songTestDSql[i])
				t.Logf("%s %s", sql, string(v))
				t.Fatalf("D%d %s <- %s <- %s", i+1, sql, string(v), string(b))
			} else if IsLog {
				t.Logf("D%d %s <- %s <- %s", i+1, sql, string(v), string(b))
			}
		}
	}
}

// 将songo格式解析为mgo的M格式
func Test_SongoParseMgo(t *testing.T) {
	// 实例1
	for i, m := range []map[string]interface{}{songTestA1, songTestA2, songTestA3} {
		if vals, err := ParseMgo(m); err != nil {
			t.Fatal(err)
		} else {
			b, _ := json.Marshal(m)
			v, _ := json.Marshal(vals)
			if string(v) != songTestAMgo[i] {
				t.Logf("%s", songTestAMgo[i])
				t.Logf("%s", string(v))
				t.Fatalf("A%d parse error %s", i+1, string(b))
			} else if IsLog {
				t.Logf("A%d %s  <- %s", i+1, string(v), string(b))
			}
		}
	}
	// 实例2
	for i, m := range []map[string]interface{}{songTestB1, songTestB2, songTestB3, songTestB4} {
		if vals, err := ParseMgo(m); err != nil {
			t.Fatal(err)
		} else {
			b, _ := json.Marshal(m)
			v, _ := json.Marshal(vals)
			if string(v) != songTestBMgo[i] {
				t.Logf("%s", songTestBMgo[i])
				t.Logf("%s", string(v))
				t.Fatalf("B%d parse error %s", i+1, string(b))
			} else if IsLog {
				t.Logf("B%d %s  <- %s", i+1, string(v), string(b))
			}
		}
	}
	// 实例3
	for i, m := range []map[string]interface{}{songTestC1, songTestC2, songTestC3, songTestC4} {
		if vals, err := ParseMgo(m); err != nil {
			t.Fatal(err)
		} else {
			b, _ := json.Marshal(m)
			v, _ := json.Marshal(vals)
			if string(v) != songTestCMgo[i] {
				t.Logf("%s", songTestCMgo[i])
				t.Logf("%s", string(v))
				t.Fatalf("C%d parse error %s", i+1, string(b))
			} else if IsLog {
				t.Logf("C%d %s  <- %s", i+1, string(v), string(b))
			}
		}
	}
	// 实例4
	for i, m := range []map[string]interface{}{songTestD1, songTestD2, songTestD3} {
		if vals, err := ParseMgo(m); err != nil {
			b, _ := json.Marshal(m)
			t.Fatalf("D%d %s %v", i+1, string(b), err)
		} else {
			b, _ := json.Marshal(m)
			v, _ := json.Marshal(vals)
			if string(v) != songTestDMgo[i] {
				t.Logf("%s", songTestDMgo[i])
				t.Logf("%s", string(v))
				t.Fatalf("D%d parse error %s", i+1, string(b))
			} else if IsLog {
				t.Logf("D%d %s <- %s", i+1, string(v), string(b))
			}
		}
	}

	// update测试
	for i, m := range []map[string]interface{}{map[string]interface{}{
		"$inc$": map[string]interface{}{
			"balance": 100,
		},
	}} {
		if vals, err := ParseMgo(m); err != nil {
			t.Fatal(err)
		} else {
			b, _ := json.Marshal(m)
			v, _ := json.Marshal(vals)
			t.Logf("D%d %s  <- %s", i+1, string(v), string(b))
		}
	}
}

// 参数过滤
func Test_SongoParseSafe(t *testing.T) {
	whiteLis := map[string]interface{}{
		"name":  true,
		"habit": true,
	}
	blackLis := map[string]interface{}{
		"uid": true,
	}
	defaultVals := map[string]interface{}{
		"country": "china",
	}
	m := map[string]interface{}{
		"uid":  "something not allow",
		"name": "jack",
		"city": "nanning",
		orm.TagQueryKeyOr: []interface{}{
			map[string]interface{}{
				orm.TagQueryKeyAnd: []interface{}{
					map[string]interface{}{
						"uid": "something not allow two",
					},
					map[string]interface{}{
						"habit": "football",
					},
				},
			},
		},
	}
	if err := ParseSafe(m, whiteLis, blackLis, defaultVals); err != nil {
		t.Fatal(err)
	} else {
		b, _ := json.Marshal(m)
		if s := string(b); s != `{"$or$":[{"$and$":[{"habit":"football"}]}],"country":"china","name":"jack"}` {
			t.Fatalf("safe map: %s", s)
		} else {
			t.Logf("safe map: %s", s)
		}
	}
}

// 搜索过滤
func Test_SongoSortSafe(t *testing.T) {
	whiteLis := []string{"createTime", "status"}
	defaultVals := []string{"-createTime"}
	input := []string{"+createTime", "updateTime"}
	if res := SortSafe(whiteLis, defaultVals, input); fmt.Sprintf("%v", res) != "[+createTime]" {
		t.Fatalf("SongoSortSafe error: %v", res)
	} else {
		t.Logf("SongoSortSafe: %v", res)
	}
}
