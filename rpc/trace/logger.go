package trace

import (
	"io"
	"os"
)

type Logger struct {
	out     io.Writer
	verbose int
}

func NewLogger(out io.Writer, verbose int) *Logger {
	if out == nil {
		out = os.Stdout
	}
	return &Logger{out: out, verbose: verbose}
}

func (p *Logger) NewTrace(verbose int, traceID, clientName, serviceMethod string) Trace {
	v := max(p.verbose, verbose)
	if v < 0 {
		return nopTrace{}
	} else if v == 0 {
		return &errorTrace{
			out:           p.out,
			traceID:       traceID,
			clientName:    clientName,
			serviceMethod: serviceMethod,
		}
	} else {
		return &verboseTrace{
			out:           p.out,
			traceID:       traceID,
			clientName:    clientName,
			serviceMethod: serviceMethod,
		}
	}
}

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}
