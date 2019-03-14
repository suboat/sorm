package orm

import "testing"

func Test_TypeJSONValue(t *testing.T) {
	var test TypeJSONValue
	//var a =[]uint8{100,188,155,255}
	if _d, err := test.Value(); err == nil {
		t.Log(_d)
	} else {
		t.Fatal(err)
	}
	//if err:=test.Scan(a);err!=nil{
	//	t.Fatalf("ssssssssssssssss %v,%v",err,test)
	//}
}
