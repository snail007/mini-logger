package logger

import (
	"fmt"
	"sync"
	"time"
)

const (
	DebugLevel = byte(1)
	InfoLevel  = byte(1 << 1)
	WarnLevel  = byte(1 << 2)
	ErrorLevel = byte(1 << 3)
	FatalLevel = byte(1 << 4)
	AllLevels  = byte(DebugLevel | InfoLevel | WarnLevel | ErrorLevel | FatalLevel)
)

type Writer interface {
	Write(e Entry)
}
type ConsoleWriter struct {
}

func (w *ConsoleWriter) Write(e Entry) {
	fmt.Println(fmt.Sprintf("[%d] [%d] %s", e.Timestamp, e.Level, e.Content))
}

type Logger struct {
	writersMap map[int][]Writer
	mu         sync.Mutex
}
type Entry struct {
	Content   string
	Timestamp int64
	Level     byte
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
	AddWriter(w Writer, levels byte)
}

// New returns a new logger
func New() MiniLogger {
	return &Logger{
		mu: sync.Mutex{},
		writersMap: map[int][]Writer{
			int(DebugLevel): []Writer{},
			int(InfoLevel):  []Writer{},
			int(WarnLevel):  []Writer{},
			int(ErrorLevel): []Writer{},
			int(FatalLevel): []Writer{},
		},
	}
}
func (l *Logger) AddWriter(w Writer, levels byte) {
	if DebugLevel&levels == DebugLevel {
		l.writersMap[int(DebugLevel)] = append(l.writersMap[int(DebugLevel)], w)
	}
	if InfoLevel&levels == InfoLevel {
		l.writersMap[int(InfoLevel)] = append(l.writersMap[int(InfoLevel)], w)

	}
	if WarnLevel&levels == WarnLevel {
		l.writersMap[int(InfoLevel)] = append(l.writersMap[int(InfoLevel)], w)

	}
	if ErrorLevel&levels == ErrorLevel {
		l.writersMap[int(ErrorLevel)] = append(l.writersMap[int(ErrorLevel)], w)

	}
	if FatalLevel&levels == FatalLevel {
		l.writersMap[int(FatalLevel)] = append(l.writersMap[int(FatalLevel)], w)
	}
}
func (l *Logger) callWriter(level byte, v ...interface{}) {
	for _, w := range l.writersMap[int(level)] {
		go w.Write(Entry{
			Timestamp: time.Now().Unix(),
			Content:   fmt.Sprint(v...),
			Level:     level,
		})
	}
}
func (l *Logger) callWriterf(level byte, format string, v ...interface{}) {
	for _, w := range l.writersMap[int(level)] {
		go w.Write(Entry{
			Timestamp: time.Now().Unix(),
			Content:   fmt.Sprintf(format, v...),
			Level:     level,
		})
	}
}
func (l *Logger) callWriterln(level byte, v ...interface{}) {
	for _, w := range l.writersMap[int(level)] {
		go w.Write(Entry{
			Timestamp: time.Now().Unix(),
			Content:   fmt.Sprintln(v...),
			Level:     level,
		})
	}
}

// Fatal works just like log.Fatal, but with a Red "FATAL[0000]" prefix.
func (l *Logger) Fatal(v ...interface{}) {
	l.callWriter(FatalLevel, v)
}

// Error works just like log.Print, but with a Red "ERROR[0000]" prefix.
func (l *Logger) Error(v ...interface{}) {
	l.callWriter(ErrorLevel, v)
}

// Warn works just like log.Print, but with a Yellow "WARN[0000]" prefix.
func (l *Logger) Warn(v ...interface{}) {
	l.callWriter(WarnLevel, v)
}

// Info works just like log.Print, but with a Blue "INFO[0000]" prefix.
func (l *Logger) Info(v ...interface{}) {
	l.callWriter(InfoLevel, v)
}

// Debug works just like log.Print, but with a Purple "DEBUG[0000]" prefix.
func (l *Logger) Debug(v ...interface{}) {
	l.callWriter(DebugLevel, v)
}

// Fatalf works just like log.Fatalf, but with a Red "FATAL[0000]" prefix.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.callWriterf(FatalLevel, format, v)
}

// Errorf works just like log.Printf, but with a Red "ERROR[0000]" prefix.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.callWriterf(ErrorLevel, format, v)
}

// Warnf works just like log.Printf, but with a Yellow "WARN[0000]" prefix.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.callWriterf(WarnLevel, format, v)
}

// Infof works just like log.Printf, but with a Blue "INFO[0000]" prefix.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.callWriterf(InfoLevel, format, v)
}

// Debugf works just like log.Printf, but with a Purple "DEBUG[0000]" prefix.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.callWriterf(DebugLevel, format, v)
}

// Fatalln works just like log.Fatalln, but with a Red "FATAL[0000]" prefix.
func (l *Logger) Fatalln(v ...interface{}) {
	l.callWriterln(FatalLevel, v)
}

// Errorln works just like log.Println, but with a Red "ERROR[0000]" prefix.
func (l *Logger) Errorln(v ...interface{}) {
	l.callWriterln(ErrorLevel, v)
}

// Warnln works just like log.Println, but with a Yellow "WARN[0000]" prefix.
func (l *Logger) Warnln(v ...interface{}) {
	l.callWriterln(WarnLevel, v)
}

// Infoln works just like log.Println, but with a Blue "INFO[0000]" prefix.
func (l *Logger) Infoln(v ...interface{}) {
	l.callWriterln(InfoLevel, v)
}

// Debugln works just like log.Println, but with a Purple "DEBUG[0000]" prefix.
func (l *Logger) Debugln(v ...interface{}) {
	l.callWriterln(DebugLevel, v)
}
