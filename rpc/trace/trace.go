package trace

import (
	"bytes"
	"io"
)

const (
	NotPrintVerbose = 0
	ErrPrintVerbose = 1
	AllPrintVerbose = 2
)

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

type Trace struct {
	out           io.Writer
	verbose       int
	traceID       string
	clientName    string
	serviceMethod string
	buffer        bytes.Buffer
}

func (t *Trace) PrintRequest(err error, body interface{}) {
}

func (t *Trace) PrintResponse(err error, body interface{}) {
}

type Logger struct {
	verbose int
}

func (p *Logger) NewTrace(verbose int, traceID, clientName, serviceMethod string) *Trace {
	return &Trace{}
}
