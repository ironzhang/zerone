package zlog

import (
	"log"
	"os"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Trace(args ...interface{})
	Tracef(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

var Default *ZLogger
var logging Logger

func init() {
	Default = NewZLogger(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile), INFO, 3)
	logging = Default
}

func SetLogger(l Logger) {
	if l == nil {
		logging = Default
	} else {
		logging = l
	}
}

func Debug(args ...interface{}) {
	logging.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logging.Debugf(format, args...)
}

func Trace(args ...interface{}) {
	logging.Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	logging.Tracef(format, args...)
}

func Info(args ...interface{}) {
	logging.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logging.Infof(format, args...)
}

func Warn(args ...interface{}) {
	logging.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logging.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logging.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logging.Errorf(format, args...)
}

func Panic(args ...interface{}) {
	logging.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	logging.Panicf(format, args...)
}

func Fatal(args ...interface{}) {
	logging.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logging.Fatalf(format, args...)
}
