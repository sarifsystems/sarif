// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

type Logger interface {
	Debugln(v ...interface{})
	Infoln(v ...interface{})
	Warnln(v ...interface{})
	Errorln(v ...interface{})
	Fatalln(v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

type defaultLogger struct{}

func (defaultLogger) Debugln(v ...interface{})               {}
func (defaultLogger) Infoln(fv ...interface{})               {}
func (defaultLogger) Warnln(fv ...interface{})               {}
func (defaultLogger) Errorln(v ...interface{})               {}
func (defaultLogger) Fatalln(v ...interface{})               {}
func (defaultLogger) Debugf(format string, v ...interface{}) {}
func (defaultLogger) Infof(format string, v ...interface{})  {}
func (defaultLogger) Warnf(format string, v ...interface{})  {}
func (defaultLogger) Errorf(format string, v ...interface{}) {}
func (defaultLogger) Fatalf(format string, v ...interface{}) {}

var defaultLog Logger = &defaultLogger{}

func SetDefaultLogger(l Logger) {
	defaultLog = l
}
