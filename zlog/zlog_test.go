package zlog_test

import (
	"fmt"
	"testing"

	"github.com/ironzhang/zerone/zlog"
)

func PrintTestZLogs(msg string) {
	fmt.Println(msg)

	zlog.Debug("debug", 1, "2", 3.0)
	zlog.Debugf("debugf: %v, %v, %v", 1, "2", 3.0)

	zlog.Trace("trace", 1, "2", 3.0)
	zlog.Tracef("tracef: %v, %v, %v", 1, "2", 3.0)

	zlog.Info("info", 1, "2", 3.0)
	zlog.Infof("infof: %v, %v, %v", 1, "2", 3.0)

	zlog.Warn("warn", 1, "2", 3.0)
	zlog.Warnf("warnf: %v, %v, %v", 1, "2", 3.0)

	zlog.Error("error", 1, "2", 3.0)
	zlog.Errorf("errorf: %v, %v, %v", 1, "2", 3.0)

	//	zlog.Panic("panic", 1, "2", 3.0)
	//	zlog.Panicf("panicf: %v, %v, %v", 1, "2", 3.0)

	//	zlog.Fatal("fatal", 1, "2", 3.0)
	//	zlog.Fatalf("fatalf: %v, %v, %v", 1, "2", 3.0)
}

func TestZlog(t *testing.T) {
	PrintTestZLogs("default")

	zlog.Default.SetLevel(zlog.DEBUG)
	PrintTestZLogs("debug")

	zlog.Default.SetLevel(zlog.TRACE)
	PrintTestZLogs("trace")

	zlog.Default.SetLevel(zlog.INFO)
	PrintTestZLogs("info")

	zlog.Default.SetLevel(zlog.WARN)
	PrintTestZLogs("warn")

	zlog.Default.SetLevel(zlog.ERROR)
	PrintTestZLogs("error")

	zlog.Default.SetLevel(zlog.FATAL)
	PrintTestZLogs("fatal")
}
