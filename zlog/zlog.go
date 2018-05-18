package zlog

import (
	"log"
	"os"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugw(message string, kvs ...interface{})

	Trace(args ...interface{})
	Tracef(format string, args ...interface{})
	Tracew(message string, kvs ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infow(message string, kvs ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnw(message string, kvs ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorw(message string, kvs ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicw(message string, kvs ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalw(message string, kvs ...interface{})
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

func Debugw(message string, kvs ...interface{}) {
	logging.Debugw(message, kvs...)
}

func Trace(args ...interface{}) {
	logging.Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	logging.Tracef(format, args...)
}

func Tracew(message string, kvs ...interface{}) {
	logging.Tracew(message, kvs...)
}

func Info(args ...interface{}) {
	logging.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logging.Infof(format, args...)
}

func Infow(message string, kvs ...interface{}) {
	logging.Infow(message, kvs...)
}

func Warn(args ...interface{}) {
	logging.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logging.Warnf(format, args...)
}

func Warnw(message string, kvs ...interface{}) {
	logging.Warnw(message, kvs...)
}

func Error(args ...interface{}) {
	logging.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logging.Errorf(format, args...)
}

func Errorw(message string, kvs ...interface{}) {
	logging.Errorw(message, kvs...)
}

func Panic(args ...interface{}) {
	logging.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	logging.Panicf(format, args...)
}

func Panicw(message string, kvs ...interface{}) {
	logging.Panicw(message, kvs...)
}

func Fatal(args ...interface{}) {
	logging.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logging.Fatalf(format, args...)
}

func Fatalw(message string, kvs ...interface{}) {
	logging.Fatalw(message, kvs...)
}
