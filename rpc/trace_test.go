package rpc

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

type TraceArgs struct {
	A, B int
}

type TraceReply struct {
	C int
}

func TestErrorTrace(t *testing.T) {
	tr := errorTrace{
		out:         ioutil.Discard,
		traceID:     "381ec868-369c-41b3-a61a-a1a1b4141c5a",
		clientName:  "TestErrorTrace",
		classMethod: "server.method",
	}
	tr.PrintRequest(TraceArgs{1, 2})
	tr.PrintResponse(nil, TraceReply{3})
	tr.PrintResponse(io.EOF, nil)
}

func TestVerboseTrace(t *testing.T) {
	tr := verboseTrace{
		out:         ioutil.Discard,
		traceID:     "6af3b859-9003-4760-a24a-fe4ab16013f0",
		clientName:  "TestVerboseTrace",
		classMethod: "server.method",
	}
	tr.PrintRequest(TraceArgs{1, 2})
	tr.PrintResponse(nil, TraceReply{3})
	tr.PrintResponse(io.EOF, nil)
}

func TestTraceLogger(t *testing.T) {
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
		lg := traceLogger{
			out:     ioutil.Discard,
			verbose: tt.loggerVerbose,
		}
		tr := lg.NewTrace("Test", tt.traceVerbose, fmt.Sprintf("%d.%d", tt.loggerVerbose, tt.traceVerbose), "TestTraceLogger", "class.method")
		tr.PrintRequest(Args{1, 2})
		tr.PrintResponse(nil, Reply{3})
		tr.PrintResponse(io.EOF, Reply{3})
	}
}
