package log

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"

	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// Ldate log date
	Ldate = log.Ldate
	// Ltime log time
	Ltime = log.Ltime
	// Lmicroseconds log microseconds
	Lmicroseconds = log.Lmicroseconds
	// Llongfile long file
	Llongfile = log.Llongfile
	// Lshortfile short file
	Lshortfile = log.Lshortfile
	// LUTC UTC
	LUTC = log.LUTC
	// LstdFlags flags
	LstdFlags = log.LstdFlags
	// PanicLevel panic level
	PanicLevel = uint32(logrus.PanicLevel)
	// FatalLevel fatal level
	FatalLevel = uint32(logrus.FatalLevel)
	// ErrorLevel error level
	ErrorLevel = uint32(logrus.ErrorLevel)
	// WarnLevel warn level
	WarnLevel = uint32(logrus.WarnLevel)
	// InfoLevel info level
	InfoLevel = uint32(logrus.InfoLevel)
	// DebugLevel debug level
	DebugLevel = uint32(logrus.DebugLevel)
)

// Log flag
var (
	// LogFlag log flag
	LogFlag = log.Llongfile //
	// Log default
	Log *Logger // default
	// RotationCount count
	RotationCount uint = 365 //
)

// Logger log struct
type Logger struct {
	*logrus.Logger
	// local
	Name         string //
	LogFlag      int    // 日志标签 非DEBUG方法
	LogFlagDebug int    // 日志 DEBUG方法
	CallerSkip   int    // 定位函数层级
}

// SetFlags log normal
func SetFlags(flg int) {
	Log.SetFlags(flg)
}

// SetLevel log normal
func SetLevel(level uint32) {
	Log.SetLevel(level)
}

// Debug debug
func Debug(args ...interface{}) {
	if Log.Level >= logrus.DebugLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Debug(args...)
	}
}

// Debugf debug
func Debugf(format string, args ...interface{}) {
	if Log.Level >= logrus.DebugLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Debugf(format, args...)
	}
}

// Debugln debug
func Debugln(args ...interface{}) {
	if Log.Level >= logrus.DebugLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Debugln(args...)
	}
}

// Info info
func Info(args ...interface{}) {
	if Log.Level >= logrus.InfoLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Info(args...)
	}
}

// Infof info
func Infof(format string, args ...interface{}) {
	if Log.Level >= logrus.InfoLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Infof(format, args...)
	}
}

// Infoln info
func Infoln(args ...interface{}) {
	if Log.Level >= logrus.InfoLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Infoln(args...)
	}
}

// Warn warn
func Warn(args ...interface{}) {
	if Log.Level >= logrus.WarnLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Warn(args...)
	}
}

// Warnf warn
func Warnf(format string, args ...interface{}) {
	if Log.Level >= logrus.WarnLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Warnf(format, args...)
	}
}

// Warnln warn
func Warnln(args ...interface{}) {
	if Log.Level >= logrus.WarnLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Warnln(args...)
	}
}

// Error error
func Error(args ...interface{}) {
	if Log.Level >= logrus.ErrorLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Error(args...)
	}
}

// Errorf error
func Errorf(format string, args ...interface{}) {
	if Log.Level >= logrus.ErrorLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Errorf(format, args...)
	}
}

// Errorln error
func Errorln(args ...interface{}) {
	if Log.Level >= logrus.ErrorLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Errorln(args...)
	}
}

// Fatal fatal
func Fatal(args ...interface{}) {
	if Log.Level >= logrus.FatalLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Fatal(args...)
	}
}

// Fatalf fatal
func Fatalf(format string, args ...interface{}) {
	if Log.Level >= logrus.FatalLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Fatalf(format, args...)
	}
}

// Fatalln fatal
func Fatalln(args ...interface{}) {
	if Log.Level >= logrus.FatalLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Fatalln(args...)
	}
}

// Panic panic
func Panic(args ...interface{}) {
	if Log.Level >= logrus.PanicLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Panic(args...)
	}
}

// Panicf panic
func Panicf(format string, args ...interface{}) {
	if Log.Level >= logrus.PanicLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Panicf(format, args...)
	}
}

// Panicln panic
func Panicln(args ...interface{}) {
	if Log.Level >= logrus.PanicLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Panicln(args...)
	}
}

// Print print
func Print(args ...interface{}) {
	logrus.NewEntry(Log.Logger).Print(args...)
}

// Printf print
func Printf(format string, args ...interface{}) {
	logrus.NewEntry(Log.Logger).Printf(format, args...)
}

// Println print
func Println(args ...interface{}) {
	logrus.NewEntry(Log.Logger).Println(args...)
}

// SetFlags log normal
func (l *Logger) SetFlags(flg int) {
	l.LogFlag |= flg
	l.LogFlagDebug = l.LogFlag
}

// SetLevel log normal
func (l *Logger) SetLevel(level uint32) {
	l.Level = logrus.Level(level)
}

// GetLevel log normal
func (l *Logger) GetLevel() uint32 {
	return uint32(l.Level)
}

// Upon up on
func (l *Logger) Upon(level uint32) (ret bool) {
	if uint32(l.Level) >= level {
		ret = true
	}
	return
}

// Debug debug
func (l *Logger) Debug(args ...interface{}) {
	if l.Level >= logrus.DebugLevel {
		l.EntryWith(l.LogFlagDebug, l.CallerSkip).Debug(args...)
	}
}

// Debugf debug
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.Level >= logrus.DebugLevel {
		l.EntryWith(l.LogFlagDebug, l.CallerSkip).Debugf(format, args...)
	}
}

// Debugln debug
func (l *Logger) Debugln(args ...interface{}) {
	if l.Level >= logrus.DebugLevel {
		l.EntryWith(l.LogFlagDebug, l.CallerSkip).Debugln(args...)
	}
}

// Info info
func (l *Logger) Info(args ...interface{}) {
	if l.Level >= logrus.InfoLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Info(args...)
	}
}

// Infof info
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.Level >= logrus.InfoLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Infof(format, args...)
	}
}

// Infoln info
func (l *Logger) Infoln(args ...interface{}) {
	if l.Level >= logrus.InfoLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Infoln(args...)
	}
}

// Warn warn
func (l *Logger) Warn(args ...interface{}) {
	if l.Level >= logrus.WarnLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Warn(args...)
	}
}

// Warnf warn
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.Level >= logrus.WarnLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Warnf(format, args...)
	}
}

// Warnln warn
func (l *Logger) Warnln(args ...interface{}) {
	if l.Level >= logrus.WarnLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Warnln(args...)
	}
}

// Error error
func (l *Logger) Error(args ...interface{}) {
	if l.Level >= logrus.ErrorLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Error(args...)
	}
}

// Errorf error
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.Level >= logrus.ErrorLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Errorf(format, args...)
	}
}

// Errorln error
func (l *Logger) Errorln(args ...interface{}) {
	if l.Level >= logrus.ErrorLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Errorln(args...)
	}
}

// Print print
func (l *Logger) Print(args ...interface{}) {
	l.EntryWith(l.LogFlag, l.CallerSkip).Print(args...)
}

// Printf print
func (l *Logger) Printf(format string, args ...interface{}) {
	l.EntryWith(l.LogFlag, l.CallerSkip).Printf(format, args...)
}

// Println print
func (l *Logger) Println(args ...interface{}) {
	l.EntryWith(l.LogFlag, l.CallerSkip).Println(args...)
}

// Fatal fatal
func (l *Logger) Fatal(args ...interface{}) {
	if l.Level >= logrus.FatalLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Fatal(args...)
	}
}

// Fatalf fatal
func (l *Logger) Fatalf(format string, args ...interface{}) {
	if l.Level >= logrus.FatalLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Fatalf(format, args...)
	}
}

// Fatalln fatal
func (l *Logger) Fatalln(args ...interface{}) {
	if l.Level >= logrus.FatalLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Fatalln(args...)
	}
}

// Panic panic
func (l *Logger) Panic(args ...interface{}) {
	if l.Level >= logrus.PanicLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Panic(args...)
	}
}

// Panicf panic
func (l *Logger) Panicf(format string, args ...interface{}) {
	if l.Level >= logrus.PanicLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Panicf(format, args...)
	}
}

// Panicln panic
func (l *Logger) Panicln(args ...interface{}) {
	if l.Level >= logrus.PanicLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Panicln(args...)
	}
}

// EntryWith 格式化输出
func (l *Logger) EntryWith(flg int, callerSkip int) *logrus.Entry {
	if flg&(log.Lshortfile|log.Llongfile) != 0 {
		if pc, file, line, ok := runtime.Caller(callerSkip); ok {
			// func
			var (
				_fnName    = runtime.FuncForPC(pc).Name()
				_fnNameDir = strings.Split(_fnName, "/")
				_fnNameLis = strings.Split(_fnName, ".")
				_fnNameSrc string
			)
			if len(_fnNameDir) > 1 {
				_fnNameSrc = _fnNameDir[0] + "/" + _fnNameDir[1] + "/"
			} else {
				_fnNameSrc = _fnNameDir[0]
			}
			fnName := _fnNameLis[len(_fnNameLis)-1]

			// file
			_pcLis := strings.Split(file, _fnNameSrc)
			filePath := strings.Join(_pcLis[1:], "")

			return l.Logger.WithFields(logrus.Fields{
				"func": fmt.Sprintf("%s:%d|%s", filePath, line, fnName),
			})
		}
	}

	return logrus.NewEntry(l.Logger)
}

// Close 关闭
func (l *Logger) Close() error {
	if l.Out != nil {
		if w, ok := l.Out.(io.WriteCloser); ok {
			return w.Close()
		}
	}
	return nil
}

// Copy 复制
func (l *Logger) Copy() (r *Logger) {
	//r = new(Logger)
	//*r = *l
	//r.Logger = logrus.New()
	//*r.Logger = *l.Logger
	//r.Logger.Out = l.Logger.Out
	r, _ = NewLog(nil)
	r.CallerSkip = l.CallerSkip
	r.SetFlags(l.LogFlag)
	r.SetLevel(uint32(l.Level))
	r.Out = l.Out
	return
}

// HookError hook
type HookError struct {
	Filepath string
	Out      io.WriteCloser
}

// Levels level
func (h *HookError) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel}
}

// Fire fire
func (h *HookError) Fire(entry *logrus.Entry) (err error) {
	if entry.Level == logrus.ErrorLevel && len(h.Filepath) > 0 {
		if _s, _err := entry.String(); _err == nil {
			if _, err := os.Stat(h.Filepath); os.IsNotExist(err) {
				if h.Out != nil {
					_ = h.Out.Close()
					h.Out = nil
				}
			}
			if h.Out == nil {
				h.Out, _ = os.OpenFile(h.Filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			}
			if h.Out != nil {
				_, _ = h.Out.Write([]byte(_s))
			}
		}
	}
	return
}

// NewLog new
func NewLog(s *Logger) (d *Logger, err error) {
	if s != nil {
		if d = s; d.Logger == nil {
			d.Logger = logrus.New()
		}
	} else {
		d = &Logger{}
		d.Logger = logrus.New()
		// default
		d.CallerSkip = 2 // 通用
		d.SetFlags(Llongfile)
		d.SetLevel(DebugLevel)
		d.Out = os.Stderr
	}
	return
}

// NewLogFile new log file
func NewLogFile(logPath string) (d *Logger) {
	var (
		//f   *os.File
		rf  *rotatelogs.RotateLogs
		err error
	)
	d, _ = NewLog(nil)

	// ensure director
	_dir := filepath.Dir(logPath)
	if _, _err := os.Stat(_dir); os.IsNotExist(_err) {
		if err = os.MkdirAll(_dir, os.ModePerm); err != nil {
			panic(err)
		}
	}

	// log file(s)
	if rf, err = rotatelogs.New(
		logPath+".%Y%m%d",
		//rotatelogs.WithLinkName(logPath),
		rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithMaxAge(-1),
		rotatelogs.WithRotationCount(RotationCount),
	); err == nil {
		d.Out = rf
	} else {
		Log.Warnln(err)
	}
	//if f, err = os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); err == nil {
	//	d.Out = f
	//}

	// hook errors
	d.AddHook(&HookError{Filepath: logPath + ".error"})

	return
}

// init
func init() {
	Log, _ = NewLog(nil)
}
