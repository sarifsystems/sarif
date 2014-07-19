package log

import (
	"log"
	"os"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
)

var Default = New()

type Logger struct {
	level LogLevel
	*log.Logger
}

func New() *Logger {
	return &Logger{
		LevelDebug,
		log.New(os.Stdout, "[stark] ", 0),
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	l.Logger.Printf("DEBUG: "+format, v...)
}

func (l *Logger) Debugln(v ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	v = append([]interface{}{"DEBUG"}, v...)
	l.Logger.Println(v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	l.Logger.Printf("INFO: "+format, v...)
}

func (l *Logger) Infoln(v ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	v = append([]interface{}{"INFO"}, v...)
	l.Logger.Println(v)
}

func (l *Logger) Warnln(v ...interface{}) {
	if l.level > LevelWarn {
		return
	}
	v = append([]interface{}{"WARN"}, v...)
	l.Logger.Println(v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level > LevelWarn {
		return
	}
	l.Logger.Printf("WARN: "+format, v...)
}
