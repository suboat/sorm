package log

import (
	"github.com/fatih/color"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"

	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
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
	PanicLevel = int(logrus.PanicLevel)
	// FatalLevel fatal level
	FatalLevel = int(logrus.FatalLevel)
	// ErrorLevel error level
	ErrorLevel = int(logrus.ErrorLevel)
	// WarnLevel warn level
	WarnLevel = int(logrus.WarnLevel)
	// InfoLevel info level
	InfoLevel = int(logrus.InfoLevel)
	// DebugLevel debug level
	DebugLevel = int(logrus.DebugLevel)
)

// Log flag
var (
	// 默认配置
	Log        *Logger                   // 默认日志
	LogOut     io.Writer = nil           // !nil: 所有日志也会同时输出到该目录
	LogFlag              = log.Llongfile // 默认输出格式
	IsInDocker bool                      // true: 当前在docker环境
	// rotate options
	RotateSuffix                 = ".%Y%m%d.part"       // 易于识别的后缀
	RotateTime     time.Duration = time.Hour * 24       //
	RotateMaxCount uint          = 365                  //
	RotateMaxAge   time.Duration = -1                   //
	RotateCompress               = gzip.BestCompression //
	//
	RegLevel = regexp.MustCompile("level=([a-z]+)") // 判断日志级别
)

// Logger log struct
type Logger struct {
	*logrus.Logger
	lock sync.RWMutex
	// local
	Out          io.Writer     // 日志输出
	Name         string        // 日志称呼
	LogFlag      int           // 输出格式 非DEBUG方法
	LogFlagDebug int           // 输出格式 DEBUG方法
	CallerSkip   int           // 定位函数层级
	Entry        *logrus.Entry //
}

//
func (l *Logger) Write(p []byte) (n int, err error) {
	if l != nil && l.Out != nil {
		// 指定输出
		if IsInDocker == false && (l.Out == os.Stdout || l.Out == os.Stderr) {
			// 终端输出
			n, err = l.color(l.Out, p)
		} else {
			// 标准输出
			n, err = l.Out.Write(p)
		}
	}
	// 全局输出
	if LogOut != nil && LogOut != l.Out {
		// 全局输出是否和日志输出相同
		if (LogOut == os.Stdout || LogOut == os.Stderr) && (l.Out == os.Stdout || l.Out == os.Stderr) {
			// 全局和指定都是终端输出
		} else {
			if IsInDocker == false && (LogOut == os.Stdout || LogOut == os.Stderr) {
				// 终端输出
				_, _ = l.color(LogOut, p)
			} else {
				// 标准输出
				_, _ = LogOut.Write(p)
			}
		}
	}
	return
}

// 终端输出
func (l *Logger) color(w io.Writer, p []byte) (n int, err error) {
	var (
		con   = string(p)
		level logrus.Level
	)
	if w == nil {
		w = color.Output
	}
	// 判断等级
	if m := RegLevel.FindStringSubmatch(con); len(m) > 1 {
		level, _ = logrus.ParseLevel(m[1])
	}
	//
	if level >= logrus.DebugLevel {
		n, err = color.New(color.FgWhite).Fprint(w, con)
	} else if level >= logrus.InfoLevel {
		n, err = color.New(color.FgGreen).Fprint(w, con)
	} else if level >= logrus.WarnLevel {
		n, err = color.New(color.FgYellow).Fprint(w, con)
	} else if level >= logrus.ErrorLevel {
		n, err = color.New(color.FgRed).Fprint(w, con)
	} else {
		n, err = color.New(color.FgRed).Fprint(w, con)
	}
	return
}

//
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.Entry != nil {
		l.Entry = l.Entry.WithField(key, value)
	} else {
		l.Entry = l.Logger.WithField(key, value)
	}
	return l
}

//
func (l *Logger) Close() (err error) {
	if l == nil && l.Out != nil {
		if w, ok := l.Out.(io.WriteCloser); ok {
			return w.Close()
		}
	}
	return
}

// SetFlags log normal
func SetFlags(flg int) {
	Log.SetFlags(flg)
}

// SetLevel log normal
func SetLevel(level int) {
	Log.SetLevel(level)
}

// Debug debug
func Debug(args ...interface{}) {
	if Log.Level >= logrus.DebugLevel {
		Log.EntryWith(Log.LogFlagDebug, Log.CallerSkip).Debug(args...)
	}
}

// Debugf debug
func Debugf(format string, args ...interface{}) {
	if Log.Level >= logrus.DebugLevel {
		Log.EntryWith(Log.LogFlagDebug, Log.CallerSkip).Debugf(format, args...)
	}
}

// Debugln debug
func Debugln(args ...interface{}) {
	if Log.Level >= logrus.DebugLevel {
		Log.EntryWith(Log.LogFlagDebug, Log.CallerSkip).Debugln(args...)
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
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Error(PubErrorConvert(args)...)
	}
}

// Errorf error
func Errorf(format string, args ...interface{}) {
	if Log.Level >= logrus.ErrorLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Errorf(format, PubErrorConvert(args)...)
	}
}

// Errorln error
func Errorln(args ...interface{}) {
	if Log.Level >= logrus.ErrorLevel {
		Log.EntryWith(Log.LogFlag, Log.CallerSkip).Errorln(PubErrorConvert(args)...)
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
	Log.EntryWith(Log.LogFlag, Log.CallerSkip).Panic(args...)
}

// Panicf panic
func Panicf(format string, args ...interface{}) {
	Log.EntryWith(Log.LogFlag, Log.CallerSkip).Panicf(format, args...)
}

// Panicln panic
func Panicln(args ...interface{}) {
	Log.EntryWith(Log.LogFlag, Log.CallerSkip).Panicln(args...)
}

// Print print
func Print(args ...interface{}) {
	Log.EntryWith(Log.LogFlag, Log.CallerSkip).Print(args...)
}

// Printf print
func Printf(format string, args ...interface{}) {
	Log.EntryWith(Log.LogFlag, Log.CallerSkip).Printf(format, args...)
}

// Println print
func Println(args ...interface{}) {
	Log.EntryWith(Log.LogFlag, Log.CallerSkip).Println(args...)
}

// SetFlags log normal
func (l *Logger) SetFlags(flg int) {
	l.LogFlag |= flg
	l.LogFlagDebug = l.LogFlag
}

// SetLevel log normal
func (l *Logger) SetLevel(level int) {
	l.Level = logrus.Level(level)
}

// GetLevel log normal
func (l *Logger) GetLevel() int {
	return int(l.Level)
}

// Upon up on
func (l *Logger) Upon(level int) (ret bool) {
	if int(l.Level) >= level {
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
		l.EntryWith(l.LogFlag, l.CallerSkip).Error(PubErrorConvert(args)...)
	}
}

// Errorf error
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.Level >= logrus.ErrorLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Errorf(format, PubErrorConvert(args)...)
	}
}

// Errorln error
func (l *Logger) Errorln(args ...interface{}) {
	if l.Level >= logrus.ErrorLevel {
		l.EntryWith(l.LogFlag, l.CallerSkip).Errorln(PubErrorConvert(args)...)
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
	l.EntryWith(l.LogFlag, l.CallerSkip).Panic(args...)
}

// Panicf panic
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.EntryWith(l.LogFlag, l.CallerSkip).Panicf(format, args...)
}

// Panicln panic
func (l *Logger) Panicln(args ...interface{}) {
	l.EntryWith(l.LogFlag, l.CallerSkip).Panicln(args...)
}

// EntryWith 格式化输出
func (l *Logger) EntryWith(flg int, callerSkip int) (ret *logrus.Entry) {
	if flg&(log.Lshortfile|log.Llongfile) != 0 {
		// func
		if pc, file, line, ok := runtime.Caller(callerSkip); ok {
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

			fs := logrus.Fields{
				"func": fmt.Sprintf("%s.%d.%s", filePath, line, fnName),
			}
			if l.Entry == nil {
				ret = l.Logger.WithFields(fs)
			} else {
				ret = l.Entry.WithFields(fs)
			}
		}
	}
	if ret == nil && l.Entry != nil {
		ret = l.Entry
	}
	if ret == nil {
		ret = logrus.NewEntry(l.Logger)
	}
	return
}

// Output
func (l *Logger) Output(callDepth int, s string) (err error) {
	// 向前延一级
	callDepth += 1
	// 从前缀解析日志级别
	if len(s) >= 3 {
		// 定制输出: nsq
		switch s[0:3] {
		case "DBG":
			if l.Level >= logrus.DebugLevel {
				l.EntryWith(l.LogFlagDebug, callDepth).Debug(s[3:])
			}
			break
		case "INF":
			if l.Level >= logrus.InfoLevel {
				l.EntryWith(l.LogFlag, callDepth).Info(s[3:])
			}
			break
		case "ERR":
			if l.Level >= logrus.ErrorLevel {
				l.EntryWith(l.LogFlag, callDepth).Error(s[3:])
			}
			break
		default:
			// 普通输出
			if l.Level >= logrus.DebugLevel {
				l.EntryWith(l.LogFlagDebug, callDepth).Debug(s[3:])
			}
		}
	} else {
		// 普通输出
		if l.Level >= logrus.DebugLevel {
			l.EntryWith(l.LogFlagDebug, callDepth).Debug(s)
		}
	}
	return
}

// Copy 复制
func (l *Logger) Copy() (r *Logger) {
	r = NewLog(nil)
	r.Out = l.Out
	r.Name = l.Name
	r.CallerSkip = l.CallerSkip
	r.SetFlags(l.LogFlag)
	r.SetLevel(int(l.Level))
	if l.Entry != nil {
		r.Entry = l.Entry.WithFields(nil)
	}
	return
}

// HookError hook
type HookError struct {
	Filepath string
	Out      io.WriteCloser
	sync     sync.RWMutex
}

// Levels level
func (h *HookError) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel}
}

// Fire err
func (h *HookError) Fire(entry *logrus.Entry) (err error) {
	if entry.Level == logrus.ErrorLevel && len(h.Filepath) > 0 {
		if _b, _err := entry.Bytes(); _err == nil {
			h.sync.Lock()
			defer h.sync.Unlock()
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
				_, _ = h.Out.Write(_b)
			}
		}
	}
	return
}

// HookRotate
type HookRotate struct {
	CompressLevel int
}

// compress
func (h *HookRotate) Handle(event rotatelogs.Event) {
	switch event.Type() {
	case rotatelogs.FileRotatedEventType:
		if h.CompressLevel <= 0 {
			return
		}
		if h.CompressLevel > 9 {
			h.CompressLevel = 9
		}
		if v, ok := event.(*rotatelogs.FileRotatedEvent); ok {
			var (
				pathPure = v.PreviousFile()
				pathGzip = pathPure + ".gz"
				filePure *os.File
				fileGzip *os.File
				writer   *gzip.Writer
				err      error
			)
			if len(pathPure) == 0 {
				return
			}
			defer func() {
				if err == nil {
					// 压缩成功:删除原文件
					_ = os.Remove(pathPure)
				} else {
					// 错误
					Errorf(`%s -> %s err: %v`, pathPure, pathGzip, err)
				}
			}()
			//
			if filePure, err = os.Open(pathPure); err != nil {
				return
			} else {
				defer filePure.Close()
			}
			if fileGzip, err = os.Create(pathGzip); err != nil {
				return
			} else {
				defer fileGzip.Close()
			}
			//
			if writer, err = gzip.NewWriterLevel(fileGzip, h.CompressLevel); err != nil {
				return
			}
			if _, err = io.Copy(writer, filePure); err != nil {
				return
			} else if err = writer.Close(); err != nil {
				return
			}
		}
		break
	default:
		break
	}
}

// NewLog new
func NewLog(s *Logger) (d *Logger) {
	if s != nil {
		if d = s; d.Logger == nil {
			d.Logger = logrus.New()
		}
	} else {
		d = &Logger{}
		d.Logger = logrus.New()
		d.Logger.Out = d
		d.Logger.SetFormatter(&logrus.TextFormatter{
			//ForceColors:     true,
			//FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.999",
		})
		// default
		d.CallerSkip = 2       // 通用
		d.SetFlags(Llongfile)  //
		d.SetLevel(DebugLevel) //
		d.Out = os.Stdout
	}
	return
}

// NewLogFile new log file
func NewLogFile(logPath string) (d *Logger) {
	var (
		err error
	)
	d = NewLog(nil)

	// 输出到文件的格式
	d.Logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})

	// ensure director
	_dir := filepath.Dir(logPath)
	if _, _err := os.Stat(_dir); os.IsNotExist(_err) {
		if err = os.MkdirAll(_dir, os.ModePerm); err != nil {
			panic(err)
		}
	}

	// name
	if ns := strings.Split(filepath.Base(logPath), "."); len(ns) > 0 {
		if len(ns) <= 2 {
			d.Name = ns[0]
		} else {
			d.Name = ns[0] + "." + ns[1]
		}
	}
	if len(d.Name) > 0 {
		d.WithField("name", d.Name)
	}

	// log file(s)
	//if f, _err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); _err == nil {
	//	d.Out = f
	//}
	if _rf, _err := rotatelogs.New(
		logPath+RotateSuffix,
		rotatelogs.WithLinkName(filepath.Base(logPath)),
		rotatelogs.WithRotationTime(RotateTime),
		rotatelogs.WithMaxAge(RotateMaxAge),
		rotatelogs.WithRotationCount(RotateMaxCount),
		rotatelogs.WithHandler(&HookRotate{CompressLevel: RotateCompress}),
	); _err == nil {
		d.Out = _rf
	} else {
		Log.Warnln(_err)
	}

	// hook errors
	d.AddHook(&HookError{Filepath: logPath + ".error"})

	return
}

// init
func init() {
	Log = NewLog(nil)
	//
	if _, err := os.Stat("/.dockerenv"); err == nil {
		IsInDocker = true
	}
}
