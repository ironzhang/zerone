package zlog

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestSprint(t *testing.T) {
	tests := []struct {
		level Level
		args  []interface{}
		str   string
	}{
		{
			level: DEBUG,
			args:  []interface{}{"1", 2, 3, "4"},
			str:   "[DEBUG] 12 34",
		},
	}
	for _, tt := range tests {
		if got, want := sprint(tt.level, tt.args...), tt.str; got != want {
			t.Errorf("sprint: %v != %v", got, want)
		}
	}
}

func TestSprintf(t *testing.T) {
	tests := []struct {
		level  Level
		format string
		args   []interface{}
		str    string
	}{
		{
			level:  DEBUG,
			format: "%v %v %v %v",
			args:   []interface{}{"1", 2, 3, "4"},
			str:    "[DEBUG] 1 2 3 4",
		},
	}
	for _, tt := range tests {
		if got, want := sprintf(tt.level, tt.format, tt.args...), tt.str; got != want {
			t.Errorf("sprint: %v != %v", got, want)
		}
	}
}

func PrintTestLogs(msg string, l *ZLogger) {
	fmt.Println(msg)

	l.Debug("debug", 1, "2", 3.0)
	l.Debugf("debugf: %v, %v, %v", 1, "2", 3.0)

	l.Trace("trace", 1, "2", 3.0)
	l.Tracef("tracef: %v, %v, %v", 1, "2", 3.0)

	l.Info("info", 1, "2", 3.0)
	l.Infof("infof: %v, %v, %v", 1, "2", 3.0)

	l.Warn("warn", 1, "2", 3.0)
	l.Warnf("warnf: %v, %v, %v", 1, "2", 3.0)

	l.Error("error", 1, "2", 3.0)
	l.Errorf("errorf: %v, %v, %v", 1, "2", 3.0)

	//	l.Fatal("fatal", 1, "2", 3.0)
	//	l.Fatalf("fatalf: %v, %v, %v", 1, "2", 3.0)
}

func TestLogger(t *testing.T) {
	var enable bool
	enable = true
	if enable {
		l := NewZLogger(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile), DEBUG, 2)

		l.SetLevel(DEBUG)
		PrintTestLogs("debug level", l)

		l.SetLevel(TRACE)
		PrintTestLogs("trace level", l)

		l.SetLevel(INFO)
		PrintTestLogs("info level", l)

		l.SetLevel(WARN)
		PrintTestLogs("warn level", l)

		l.SetLevel(ERROR)
		PrintTestLogs("error level", l)

		l.SetLevel(FATAL)
		PrintTestLogs("fatal level", l)
	}
}
