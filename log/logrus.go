package log

import (
	"github.com/Sirupsen/logrus"

	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

// Log flag
var (
	LogFlag int

	Log = logrus.New()
)

// log normal
func SetFlags(flg int) {
	LogFlag |= flg
}

// init
func init() {
	//logrus.SetOutput(os.Stderr)
	Log.Out = os.Stderr
	// test
	SetFlags(log.Llongfile)
	//SetLevel(logrus.DebugLevel)
	SetLevel(logrus.InfoLevel)
}

//
func EntryWith(flg int) *logrus.Entry {
	if flg&(log.Lshortfile|log.Llongfile) != 0 {
		if pc, file, line, ok := runtime.Caller(2); ok {
			// func
			_fnName := runtime.FuncForPC(pc).Name()
			_fnNameLis := strings.Split(_fnName, ".")
			_fnNameSrc := strings.Split(_fnName, "/")[0]
			fnName := _fnNameLis[len(_fnNameLis)-1]

			// file
			_pcLis := strings.Split(file, _fnNameSrc)
			filePath := _fnNameSrc + strings.Join(_pcLis[1:], "")

			return Log.WithFields(logrus.Fields{
				"func": fmt.Sprintf("%s:%d|%s", filePath, line, fnName),
			})
		} else {
			return logrus.NewEntry(Log)
		}
	} else {
		return logrus.NewEntry(Log)
	}
}

//
func SetLevel(level logrus.Level) {
	Log.Level = level
}

// debug
func Debug(args ...interface{}) {
	//logrus.Debug(args...)
	if Log.Level >= logrus.DebugLevel {
		EntryWith(LogFlag).Debug(args...)
	}
}

// info
func Info(args ...interface{}) {
	//logrus.Info(args...)
	if Log.Level >= logrus.InfoLevel {
		EntryWith(LogFlag).Info(args...)
	}
}

// warn
func Warn(args ...interface{}) {
	//logrus.Warn(args...)
	if Log.Level >= logrus.WarnLevel {
		EntryWith(LogFlag).Warn(args...)
	}
}

// error
func Error(args ...interface{}) {
	//logrus.Error(args...)
	if Log.Level >= logrus.ErrorLevel {
		EntryWith(LogFlag).Error(args...)
	}
}

// fatal
func Fatal(args ...interface{}) {
	//logrus.Fatal(args...)
	if Log.Level >= logrus.FatalLevel {
		EntryWith(LogFlag).Fatal(args...)
	}
}

// panic
func Panic(args ...interface{}) {
	//logrus.Panic(args...)
	if Log.Level >= logrus.PanicLevel {
		EntryWith(LogFlag).Panic(args...)
	}
}
