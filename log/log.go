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
	LevelError
	LevelFatal
	LevelCritical
)

var Default = New()

type Logger struct {
	level LogLevel
	*log.Logger
}

func New() *Logger {
	return &Logger{
		LevelDebug,
		log.New(os.Stdout, "", log.LstdFlags),
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
	l.Logger.Println(v...)
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

func (l *Logger) Errorln(v ...interface{}) {
	if l.level > LevelError {
		return
	}
	v = append([]interface{}{"ERROR"}, v...)
	l.Logger.Println(v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level > LevelError {
		return
	}
	l.Logger.Printf("ERROR: "+format, v...)
}

func (l *Logger) Fatalln(v ...interface{}) {
	if l.level > LevelFatal {
		return
	}
	v = append([]interface{}{"FATAL"}, v...)
	l.Logger.Fatalln(v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.level > LevelFatal {
		return
	}
	l.Logger.Fatalf("FATAL: "+format, v...)
}

func (l *Logger) Criticalln(v ...interface{}) {
	if l.level > LevelCritical {
		return
	}
	v = append([]interface{}{"CRITICAL"}, v...)
	l.Logger.Panicln(v...)
}

func (l *Logger) Criticalf(format string, v ...interface{}) {
	if l.level > LevelCritical {
		return
	}
	l.Logger.Panicf("CRITICAL: "+format, v...)
}
