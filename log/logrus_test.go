package log

import (
	"os"
	"testing"
	"time"
)

//
func Test_File(t *testing.T) {
	RotateTime = time.Second
	RotateSuffix = ".%Y%m%d%H%M%S"
	//
	var (
		a *os.File
		l = NewLogFile("./log/test.log")
		//l  = NewLog(nil)
		tk = time.NewTicker(time.Second)
	)
	if false {
		a, _ = os.OpenFile("./all.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		LogOut = a
	} else {
		LogOut = os.Stderr
	}

	l.Infof("begin")
	go func() {
		i := 0
		for {
			select {
			case <-tk.C:
				i += 1
				if i%3 == 0 {
					l.Errorf("%d", i)
				} else if i%2 == 0 {
					l.Copy().Infof("%d", i)
				} else {
					l.Debugf("%d", i)
				}

			}
		}
	}()

	Debug("sleep 1")
	Log.Debug("sleep 2")
	time.Sleep(time.Second * 7)
	l.Errorf("end")
}
