package log

var (
	HookErrorConvert func(interface{}) interface{} = nil
)

// PubErrorConvert 错误转换
func PubErrorConvert(in []interface{}) (out []interface{}) {
	if HookErrorConvert == nil {
		return in
	}
	for _, v := range in {
		if v != nil {
			v = HookErrorConvert(v)
		}
		out = append(out, v)
	}
	return
}
