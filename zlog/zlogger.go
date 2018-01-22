package zlog

import (
	"fmt"
	"log"
	"os"
)

type Level int

const (
	DEBUG Level = -2
	TRACE Level = -1
	INFO  Level = 0
	WARN  Level = 1
	ERROR Level = 2
	PANIC Level = 3
	FATAL Level = 4
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case PANIC:
		return "PANIC"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type ZLogger struct {
	logger    *log.Logger
	level     Level
	calldepth int
}

func NewZLogger(logger *log.Logger, level Level, calldepth int) *ZLogger {
	return &ZLogger{
		logger:    logger,
		level:     level,
		calldepth: calldepth,
	}
}

func (p *ZLogger) SetLogger(l *log.Logger) {
	p.logger = l
}

func (p *ZLogger) SetLevel(l Level) {
	p.level = l
}

func (p *ZLogger) SetCalldepth(calldepth int) {
	p.calldepth = calldepth
}

func (p *ZLogger) Debug(args ...interface{}) {
	if p.level <= DEBUG {
		p.logger.Output(p.calldepth, sprint(DEBUG, args...))
	}
}

func (p *ZLogger) Debugf(format string, args ...interface{}) {
	if p.level <= DEBUG {
		p.logger.Output(p.calldepth, sprintf(DEBUG, format, args...))
	}
}

func (p *ZLogger) Trace(args ...interface{}) {
	if p.level <= TRACE {
		p.logger.Output(p.calldepth, sprint(TRACE, args...))
	}
}

func (p *ZLogger) Tracef(format string, args ...interface{}) {
	if p.level <= TRACE {
		p.logger.Output(p.calldepth, sprintf(TRACE, format, args...))
	}
}

func (p *ZLogger) Info(args ...interface{}) {
	if p.level <= INFO {
		p.logger.Output(p.calldepth, sprint(INFO, args...))
	}
}

func (p *ZLogger) Infof(format string, args ...interface{}) {
	if p.level <= INFO {
		p.logger.Output(p.calldepth, sprintf(INFO, format, args...))
	}
}

func (p *ZLogger) Warn(args ...interface{}) {
	if p.level <= WARN {
		p.logger.Output(p.calldepth, sprint(WARN, args...))
	}
}

func (p *ZLogger) Warnf(format string, args ...interface{}) {
	if p.level <= WARN {
		p.logger.Output(p.calldepth, sprintf(WARN, format, args...))
	}
}

func (p *ZLogger) Error(args ...interface{}) {
	if p.level <= ERROR {
		p.logger.Output(p.calldepth, sprint(ERROR, args...))
	}
}

func (p *ZLogger) Errorf(format string, args ...interface{}) {
	if p.level <= ERROR {
		p.logger.Output(p.calldepth, sprintf(ERROR, format, args...))
	}
}

func (p *ZLogger) Panic(args ...interface{}) {
	if p.level <= PANIC {
		p.logger.Output(p.calldepth, sprint(PANIC, args...))
	}
	panic(fmt.Sprint(args...))
}

func (p *ZLogger) Panicf(format string, args ...interface{}) {
	if p.level <= PANIC {
		p.logger.Output(p.calldepth, sprintf(PANIC, format, args...))
	}
	panic(fmt.Sprintf(format, args...))
}

func (p *ZLogger) Fatal(args ...interface{}) {
	if p.level <= FATAL {
		p.logger.Output(p.calldepth, sprint(FATAL, args...))
	}
	os.Exit(1)
}

func (p *ZLogger) Fatalf(format string, args ...interface{}) {
	if p.level <= FATAL {
		p.logger.Output(p.calldepth, sprintf(FATAL, format, args...))
	}
	os.Exit(1)
}

func sprint(l Level, args ...interface{}) string {
	return "[" + l.String() + "] " + fmt.Sprint(args...)
}

func sprintf(l Level, format string, args ...interface{}) string {
	return "[" + l.String() + "] " + fmt.Sprintf(format, args...)
}
