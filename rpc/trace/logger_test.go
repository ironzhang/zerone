package trace

import (
	"reflect"
	"testing"
)

func TestLogger(t *testing.T) {
	const (
		server      = true
		traceID     = "TraceID"
		clientName  = "ClientName"
		clientAddr  = "ClientAddr"
		serverName  = "ServerName"
		serverAddr  = "ServerAddr"
		classMethod = "ClassMethod"
	)

	tests := []struct {
		out           Output
		loggerVerbose int
		traceVerbose  int
		trace         Trace
	}{
		{
			out:   nil,
			trace: nopTrace{},
		},

		{
			out:           DefaultOutput,
			loggerVerbose: -1,
			traceVerbose:  -1,
			trace:         nopTrace{},
		},
		{
			out:           DefaultOutput,
			loggerVerbose: -1,
			traceVerbose:  0,
			trace: &errTrace{
				out:         DefaultOutput,
				server:      server,
				traceID:     traceID,
				clientName:  clientName,
				clientAddr:  clientAddr,
				serverName:  serverName,
				serverAddr:  serverAddr,
				classMethod: classMethod,
			},
		},
		{
			out:           DefaultOutput,
			loggerVerbose: -1,
			traceVerbose:  1,
			trace: &verboseTrace{
				out:         DefaultOutput,
				server:      server,
				traceID:     traceID,
				clientName:  clientName,
				clientAddr:  clientAddr,
				serverName:  serverName,
				serverAddr:  serverAddr,
				classMethod: classMethod,
			},
		},

		{
			out:           DefaultOutput,
			loggerVerbose: 0,
			traceVerbose:  -1,
			trace: &errTrace{
				out:         DefaultOutput,
				server:      server,
				traceID:     traceID,
				clientName:  clientName,
				clientAddr:  clientAddr,
				serverName:  serverName,
				serverAddr:  serverAddr,
				classMethod: classMethod,
			},
		},
		{
			out:           DefaultOutput,
			loggerVerbose: 0,
			traceVerbose:  0,
			trace: &errTrace{
				out:         DefaultOutput,
				server:      server,
				traceID:     traceID,
				clientName:  clientName,
				clientAddr:  clientAddr,
				serverName:  serverName,
				serverAddr:  serverAddr,
				classMethod: classMethod,
			},
		},
		{
			out:           DefaultOutput,
			loggerVerbose: 0,
			traceVerbose:  1,
			trace: &verboseTrace{
				out:         DefaultOutput,
				server:      server,
				traceID:     traceID,
				clientName:  clientName,
				clientAddr:  clientAddr,
				serverName:  serverName,
				serverAddr:  serverAddr,
				classMethod: classMethod,
			},
		},

		{
			out:           DefaultOutput,
			loggerVerbose: 1,
			traceVerbose:  -1,
			trace: &verboseTrace{
				out:         DefaultOutput,
				server:      server,
				traceID:     traceID,
				clientName:  clientName,
				clientAddr:  clientAddr,
				serverName:  serverName,
				serverAddr:  serverAddr,
				classMethod: classMethod,
			},
		},
		{
			out:           DefaultOutput,
			loggerVerbose: 1,
			traceVerbose:  0,
			trace: &verboseTrace{
				out:         DefaultOutput,
				server:      server,
				traceID:     traceID,
				clientName:  clientName,
				clientAddr:  clientAddr,
				serverName:  serverName,
				serverAddr:  serverAddr,
				classMethod: classMethod,
			},
		},
		{
			out:           DefaultOutput,
			loggerVerbose: 1,
			traceVerbose:  1,
			trace: &verboseTrace{
				out:         DefaultOutput,
				server:      server,
				traceID:     traceID,
				clientName:  clientName,
				clientAddr:  clientAddr,
				serverName:  serverName,
				serverAddr:  serverAddr,
				classMethod: classMethod,
			},
		},
	}
	for i, tt := range tests {
		l := NewLogger()
		l.SetOutput(tt.out)
		l.SetVerbose(tt.loggerVerbose)
		trace := l.NewTrace(server, tt.traceVerbose, traceID, clientName, clientAddr, serverName, serverAddr, classMethod)
		if got, want := trace, tt.trace; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: trace: got %#v, want %#v", i, got, want)
		} else {
			t.Logf("%d: trace: got %#v", i, got)
		}

		//trace.Request(nil)
		//trace.Response(nil, nil)
	}
}
