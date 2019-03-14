package log

import "testing"

func Test_Log(t *testing.T) {
	log, err := NewLog(nil)
	if err != nil {
		// err 处理
		Error(err)
		Errorf("%v", err)
		Errorln(err)
		Panic(err)
	} else {
		// Debug
		Debug(log)
		Debugf("%v", log)
		Debugln(log)
		// Info
		Info(log)
		Infof("%v", log)
		Infoln(log)
		// print
		Print(log)
		Printf("%v", log)
		Println(log)
		// warn
		Warn(log)
		Warnf("%v", log)
		Warnln(log)
	}

	SetLevel(10)
	SetFlags(10)
	// err 处理方式
	if err != nil {
		Panicf("%v", err)
	}
	if err != nil {
		Panicln(err)
	}
	if err != nil {
		Fatal(err)
	}
	if err != nil {
		Fatalf("%v", err)
	}
	if err != nil {
		Fatalln(err)
	}

}
