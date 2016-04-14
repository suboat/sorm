package orm

import (
	"database/sql/driver"
	"encoding/json"
)

// 索引信息
type Index map[string]interface{}

// 模型定义
type Model interface {
	String() string                // 表名
	Objects() Objects              // 返回object对象
	NewUid() Uid                   // 返回一个新的唯一的ID
	UidValid(Uid) error            // 判断Uid是否合法
	EnsureIndex(index Index) error // 确认索引
	// 通过tag来定义索引:
	// unique索引    Name string `sorm:"unique"`
	// index索引     Name string `sorm:"index"`
	// TODO: 全文索引 Name string `sorm:"text"`
	EnsureIndexWithTag(st interface{}) error // 通过struct的tag来添加/确认索引
	// 读写锁
	Lock()
	Unlock()
	RLock()
	RUnlock()
	// 事务
	Begin() (Trans, error) // 事务开始
	Rollback(Trans) error  // 回滚操作
	Commit(Trans) error    // 阶段二提交
	AutoTrans(Trans) error // 自动回滚或提交
}

// 通用json内联字段
type TypeJsonValue struct{}

func (m TypeJsonValue) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *TypeJsonValue) Scan(src interface{}) (err error) {
	return json.Unmarshal(src.([]byte), m)
}
