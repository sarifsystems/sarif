// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"log"
	"os"
)

type LogLogLevel int

const (
	LogLevelDebug LogLogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelCritical
)

var DefaultLog = New(
	LogLevelDebug,
	log.New(os.Stderr, "", log.LstdFlags),
)

type Logger struct {
	level LogLogLevel
	*log.Logger
}

func New(level LogLogLevel, l *log.Logger) *Logger {
	return &Logger{
		level,
		l,
	}
}

func (l *Logger) SetLevel(level LogLogLevel) {
	l.level = level
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level > LogLevelDebug {
		return
	}
	l.Logger.Printf("DEBUG "+format, v...)
}

func (l *Logger) Debugln(v ...interface{}) {
	if l.level > LogLevelDebug {
		return
	}
	v = append([]interface{}{"DEBUG"}, v...)
	l.Logger.Println(v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level > LogLevelInfo {
		return
	}
	l.Logger.Printf("INFO "+format, v...)
}

func (l *Logger) Infoln(v ...interface{}) {
	if l.level > LogLevelInfo {
		return
	}
	v = append([]interface{}{"INFO"}, v...)
	l.Logger.Println(v...)
}

func (l *Logger) Warnln(v ...interface{}) {
	if l.level > LogLevelWarn {
		return
	}
	v = append([]interface{}{"WARN"}, v...)
	l.Logger.Println(v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level > LogLevelWarn {
		return
	}
	l.Logger.Printf("WARN "+format, v...)
}

func (l *Logger) Errorln(v ...interface{}) {
	if l.level > LogLevelError {
		return
	}
	v = append([]interface{}{"ERROR"}, v...)
	l.Logger.Println(v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level > LogLevelError {
		return
	}
	l.Logger.Printf("ERROR "+format, v...)
}

func (l *Logger) Fatalln(v ...interface{}) {
	if l.level > LogLevelFatal {
		return
	}
	v = append([]interface{}{"FATAL"}, v...)
	l.Logger.Fatalln(v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.level > LogLevelFatal {
		return
	}
	l.Logger.Fatalf("FATAL "+format, v...)
}

func (l *Logger) Criticalln(v ...interface{}) {
	if l.level > LogLevelCritical {
		return
	}
	v = append([]interface{}{"CRITICAL"}, v...)
	l.Logger.Panicln(v...)
}

func (l *Logger) Criticalf(format string, v ...interface{}) {
	if l.level > LogLevelCritical {
		return
	}
	l.Logger.Panicf("CRITICAL "+format, v...)
}
