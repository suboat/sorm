package orm

import (
	"errors"
	"testing"
)

func Test_M(t *testing.T) {
	m := NewM()
	// 建立map
	m.Set("name", "Tom")
	if "Tom" != m.Get("name") && m.Hav("name") {
		t.Errorf("%v", errors.New("writer and read is diff"))
		return
	}
	// 改名字
	m.SetExist("name", "Jack")
	if "Jack" != m.Get("name") && m.Hav("name") {
		t.Errorf("%v", errors.New("writer and read is diff"))
		return
	}
	// 建立新的key和value,如果不存在建立成功
	m.SetNotExist("age", 18)
	if 18 != m.Get("age") && m.Hav("age") {
		t.Errorf("%v", errors.New("writer and read is diff"))
		return
	}
	// 更新
	if err := m.Update(&M{"age": 20}); err != nil {
		t.Error(err)
		return
	}
	if 20 != m.Get("age") && m.Hav("age") {
		t.Errorf("%v", errors.New("writer and read is diff"))
		return
	}
	// 判断键值是否存在
	if m.IsEmpty() {
		t.Errorf("%v", errors.New("writer value ,but read is nil"))
		return
	}
	if len(m.GetString("name")) == 0 {
		t.Errorf("%v", errors.New("writer value ,but read is nil"))
		return
	}
	if m.Map() == nil {
		t.Errorf("%v", errors.New("writer value ,but read is nil"))
		return
	}
	t.Log(m.Scrape(map[string]interface{}{"name": "Jack"}, map[string]interface{}{}, map[string]interface{}{}))

}
