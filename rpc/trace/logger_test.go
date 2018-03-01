package trace

import (
	"fmt"
	"io"
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		loggerVerbose int
		traceVerbose  int
	}{
		{
			loggerVerbose: -1,
			traceVerbose:  -1,
		},
		{
			loggerVerbose: -1,
			traceVerbose:  0,
		},
		{
			loggerVerbose: -1,
			traceVerbose:  1,
		},
		{
			loggerVerbose: 0,
			traceVerbose:  -1,
		},
		{
			loggerVerbose: 0,
			traceVerbose:  0,
		},
		{
			loggerVerbose: 0,
			traceVerbose:  1,
		},
		{
			loggerVerbose: 1,
			traceVerbose:  -1,
		},
		{
			loggerVerbose: 1,
			traceVerbose:  0,
		},
		{
			loggerVerbose: 1,
			traceVerbose:  1,
		},
	}
	for _, tt := range tests {
		log := NewLogger(nil, tt.loggerVerbose)
		tr := log.NewTrace(tt.traceVerbose, fmt.Sprintf("%d.%d", tt.loggerVerbose, tt.traceVerbose), "client", "service.method")
		tr.PrintRequest(Args{1, 2})
		tr.PrintResponse(nil, Reply{3})
		tr.PrintResponse(io.EOF, Reply{3})
	}
}
