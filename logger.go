package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	DebugLevel uint8 = 1
	InfoLevel  uint8 = 1 << 1
	WarnLevel  uint8 = 1 << 2
	ErrorLevel uint8 = 1 << 3
	FatalLevel uint8 = 1 << 4
	AllLevels  uint8 = byte(DebugLevel | InfoLevel | WarnLevel | ErrorLevel | FatalLevel)
)

var (
	flush = &sync.WaitGroup{}
	exit  = false
)

type Writer interface {
	Write(e Entry)
}

type Logger struct {
	writersMap map[int][]*WrappedWriter
	mu         sync.Mutex
	modeSafe   bool
	wait       sync.WaitGroup
}
type WrappedWriter struct {
	lock   *sync.Mutex
	chn    chan Entry
	writer Writer
	wait   sync.WaitGroup
}
type Entry struct {
	Content               string
	Timestamp             int64
	Milliseconds          int64
	TimestampMilliseconds int64
	Level                 uint8
}
type MiniLogger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Debugln(v ...interface{})
	Infoln(v ...interface{})
	Warnln(v ...interface{})
	Errorln(v ...interface{})
	Fatalln(v ...interface{})
	AddWriter(w Writer, levels byte) MiniLogger
	Safe() MiniLogger
	Unsafe() MiniLogger
}

func (e *Entry) getLevelString() string {
	switch e.Level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO "
	case WarnLevel:
		return "WARN "
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	}
	return "UNKOWN"
}

// New returns a new logger
//modeSafe when true : must call logger.Flush() in main defer
//if not do that, may be lost message,but this mode has a highest performance
//when false : each message can be processed , but this mode may be has a little lower performance
//beacuse of logger must wait for  all writers process done with each messsage.
//note: if you do logging before call os.Exit(), you had better to set safeMode to true before call os.Exit().
func New(modeSafe bool) MiniLogger {
	return &Logger{
		mu:       sync.Mutex{},
		modeSafe: modeSafe,
		wait:     sync.WaitGroup{},
		writersMap: map[int][]*WrappedWriter{
			int(DebugLevel): []*WrappedWriter{},
			int(InfoLevel):  []*WrappedWriter{},
			int(WarnLevel):  []*WrappedWriter{},
			int(ErrorLevel): []*WrappedWriter{},
			int(FatalLevel): []*WrappedWriter{},
		},
	}
}

//Flush wait for process all left entry
func Flush() {
	exit = true
	flush.Wait()
}

//Safe : if you call Safe() , then you must call logger.Flush() in main defer
//if not do that, may be lost message,but this mode has a highest performance
//note: if you do logging before call os.Exit(), you had better to call Safe() before call os.Exit().
//you call call func Safe() or Unsafe() in any where any time to switch safe mode.
func (l *Logger) Safe() MiniLogger {
	l.modeSafe = true
	return l
}

//Unsafe : if you call Unsafe(), each message can be processed immediately,
//but this mode may be has  lower performance,beacuse of logger must wait for
//all writers process done with each messsage.
//note: if you do logging before call os.Exit(), you had better to call Safe() before call os.Exit().
//you call call func Safe() or Unsafe() in any where any time to switch safe mode.
func (l *Logger) Unsafe() MiniLogger {
	l.modeSafe = false
	return l
}
func (l *Logger) AddWriter(writer Writer, levels byte) MiniLogger {
	w := &WrappedWriter{
		lock:   &sync.Mutex{},
		writer: writer,
		chn:    make(chan Entry, 1024),
	}
	if DebugLevel&levels == DebugLevel {
		l.writersMap[int(DebugLevel)] = append(l.writersMap[int(DebugLevel)], w)
	}
	if InfoLevel&levels == InfoLevel {
		l.writersMap[int(InfoLevel)] = append(l.writersMap[int(InfoLevel)], w)
	}
	if WarnLevel&levels == WarnLevel {
		l.writersMap[int(WarnLevel)] = append(l.writersMap[int(WarnLevel)], w)
	}
	if ErrorLevel&levels == ErrorLevel {
		l.writersMap[int(ErrorLevel)] = append(l.writersMap[int(ErrorLevel)], w)
	}
	if FatalLevel&levels == FatalLevel {
		l.writersMap[int(FatalLevel)] = append(l.writersMap[int(FatalLevel)], w)
	}
	flush.Add(1)
	go func() {
		defer func() {
			flush.Done()
		}()
		for {
			select {
			case entry, ok := <-w.chn:
				if ok {
					if l.modeSafe {
						l.wait.Done()
					}
					w.writer.Write(entry)
					if entry.Level == FatalLevel {
						os.Exit(0)
					}
				} else {
					return
				}
			default:
				if exit {
					return
				}
			}
		}
	}()
	return l
}
func (l *Logger) callWriter(level byte, t, foramt string, v ...interface{}) {
	c := ""
	if t == "f" && len(v) == 0 {
		v = append(v, foramt)
		foramt = "%s"
	}
	switch t {
	case "ln":
		c = fmt.Sprintln(v...)
	case "f":
		c = fmt.Sprintf(foramt, v...)
	default:
		c = fmt.Sprint(v...)
	}
	for _, w := range l.writersMap[int(level)] {
		now := time.Now().UnixNano()
		nowUnix := time.Now().Unix()
		mili := (now / 1000000) - nowUnix*1000
		if l.modeSafe {
			l.wait.Add(1)
		}
		w.chn <- Entry{
			Timestamp:             nowUnix,
			TimestampMilliseconds: now / 10000000,
			Milliseconds:          mili,
			Content:               c,
			Level:                 level,
		}
	}
	if l.modeSafe {
		l.wait.Wait()
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	l.callWriter(FatalLevel, "", "", v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.callWriter(ErrorLevel, "", "", v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.callWriter(WarnLevel, "", "", v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.callWriter(InfoLevel, "", "", v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.callWriter(DebugLevel, "", "", v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.callWriter(FatalLevel, "f", format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.callWriter(ErrorLevel, "f", format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.callWriter(WarnLevel, "f", format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.callWriter(InfoLevel, "f", format, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.callWriter(DebugLevel, "f", format, v...)
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.callWriter(FatalLevel, "ln", "", v...)
}

func (l *Logger) Errorln(v ...interface{}) {
	l.callWriter(ErrorLevel, "ln", "", v...)
}

func (l *Logger) Warnln(v ...interface{}) {
	l.callWriter(WarnLevel, "ln", "", v...)
}

func (l *Logger) Infoln(v ...interface{}) {
	l.callWriter(InfoLevel, "ln", "", v...)
}

func (l *Logger) Debugln(v ...interface{}) {
	l.callWriter(DebugLevel, "ln", "", v...)
}
