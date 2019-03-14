# 查询数据

> 使用`map[string]interface{}`类型作为检索输入参数。  
> 语法参见[songo白皮书](https://github.com/suboat/songo)。

## 快速入门

 简单的查询:

 ```go
// query
func demoQueryData(m orm.Model) (ret []*demoPerson, meta *orm.Meta, err error) {
	// records
	ret = []*demoPerson{}
	// query
	q := m.Objects().
		Filter(orm.M{"age" + orm.TagValGte: 16}). // select age >= 16
		Sort("-age").                             // sort by age
		Skip(5).                                  // skip 5 records
		Limit(20)                                 // limit 20 records
	if err = q.All(&ret); err != nil {
		return
	} else if meta, err = q.Meta(); err != nil {
		return
	}
	return
}
 ```