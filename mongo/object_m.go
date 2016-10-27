package mongo

import (
	"github.com/suboat/sorm"
	"github.com/suboat/sorm/log"
	"gopkg.in/mgo.v2/bson"
)

var (
	ParserMapMax = 2 // 目前只支持两层解析
)

// 解析为标准搜索为mgo的搜索
func ParserM(s orm.M) (d bson.M, err error) {
	var (
		m  map[string]interface{}  = s
		mp *map[string]interface{} = nil
	)
	if mp, err = parserM(&m, 0); err != nil {
		return
	}
	d = *mp
	return
}
func ParserMMust(s orm.M) (d bson.M) {
	var err error
	if d, err = ParserM(s); err != nil {
		log.Error(err)
		d = bson.M(s)
	}
	return
}

func parserM(s *map[string]interface{}, deep int) (d *map[string]interface{}, err error) {
	if s == nil {
		err = orm.ErrMUndefined
		return
	} else if deep >= ParserMapMax {
		return
	} else {
		d = s
	}
	for k, v := range *s {
		kFix := k
		switch k {
		case orm.TagQueryKeyOr:
			kFix = "$or"
			break
		case orm.TagQueryKeyAnd:
			kFix = "$and"
			break
		case orm.TagQueryKeyIn:
			kFix = "$in"
			break
		default:
			break
		}
		if _s, _ok := v.(string); _ok == true {
			if _tag, _val := orm.IsTagValAuto(_s); len(_tag) > 0 {
				// 转义
				switch _tag {
				case orm.TagValNo, orm.TagValNe:
					_tag = "$ne"
					break
				case orm.TagValLike:
					_tag = "$regex"
					if _v, _ok := _val.(string); _ok == true {
						_val = bson.RegEx{_v, "i"}
					} else {
						// not support method
					}
					break
				case orm.TagValLt:
					_tag = "$lt"
					break
				case orm.TagValLte:
					_tag = "$lte"
					break
				case orm.TagValGt:
					_tag = "$gt"
					break
				case orm.TagValGte:
					_tag = "$gte"
					break
				default:
					break
				}
				v = map[string]interface{}{
					_tag: _val,
				}
			}
		} else if _m, _ok := v.(map[string]interface{}); _ok == true {
			// 如果是map再继续解析
			if mRec, _err := parserM(&_m, deep+1); _err != nil {
				err = _err
				return
			} else {
				v = *mRec
			}
		}
		// 转义
		(*s)[kFix] = v
		if kFix != k {
			delete((*s), k)
		}
	}
	return
}
