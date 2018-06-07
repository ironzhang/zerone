package trace

import (
	"io"
	"reflect"
	"testing"
	"time"
)

func init() {
	t := time.Now()
	timeNow = func() time.Time {
		return t
	}
}

type TestOutput struct {
	request  Request
	response Response
}

func (p *TestOutput) Request(r Request) {
	p.request = r
}

func (p *TestOutput) Response(r Response) {
	p.response = r
}

func TestErrTrace(t *testing.T) {
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
		args     interface{}
		err      error
		reply    interface{}
		request  Request
		response Response
	}{
		{
			args:     nil,
			reply:    nil,
			request:  Request{},
			response: Response{},
		},
		{
			args: 1,
			err:  io.EOF,
			request: Request{
				Server:      server,
				Start:       timeNow(),
				TraceID:     traceID,
				ClientName:  clientName,
				ClientAddr:  clientAddr,
				ServerName:  serverName,
				ServerAddr:  serverAddr,
				ClassMethod: classMethod,
				Args:        1,
			},
			response: Response{
				Server:      server,
				Start:       timeNow(),
				End:         timeNow(),
				TraceID:     traceID,
				ClientName:  clientName,
				ClientAddr:  clientAddr,
				ServerName:  serverName,
				ServerAddr:  serverAddr,
				ClassMethod: classMethod,
				Error:       io.EOF,
			},
		},
	}
	for i, tt := range tests {
		out := &TestOutput{}
		trace := errTrace{
			out:         out,
			server:      server,
			traceID:     traceID,
			clientName:  clientName,
			clientAddr:  clientAddr,
			serverName:  serverName,
			serverAddr:  serverAddr,
			classMethod: classMethod,
		}
		trace.Request(tt.args)
		trace.Response(tt.err, tt.reply)
		if got, want := out.request, tt.request; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: request: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: request: got %v", i, got)
		}
		if got, want := out.response, tt.response; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: response: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: response: got %v", i, got)
		}
	}
}

func TestVerboseTrace(t *testing.T) {
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
		args     interface{}
		err      error
		reply    interface{}
		request  Request
		response Response
	}{
		{
			args:  nil,
			reply: "hello",
			request: Request{
				Server:      server,
				Start:       timeNow(),
				TraceID:     traceID,
				ClientName:  clientName,
				ClientAddr:  clientAddr,
				ServerName:  serverName,
				ServerAddr:  serverAddr,
				ClassMethod: classMethod,
				Args:        nil,
			},
			response: Response{
				Server:      server,
				Start:       timeNow(),
				End:         timeNow(),
				TraceID:     traceID,
				ClientName:  clientName,
				ClientAddr:  clientAddr,
				ServerName:  serverName,
				ServerAddr:  serverAddr,
				ClassMethod: classMethod,
				Reply:       "hello",
			},
		},
		{
			args: 1,
			err:  io.EOF,
			request: Request{
				Server:      server,
				Start:       timeNow(),
				TraceID:     traceID,
				ClientName:  clientName,
				ClientAddr:  clientAddr,
				ServerName:  serverName,
				ServerAddr:  serverAddr,
				ClassMethod: classMethod,
				Args:        1,
			},
			response: Response{
				Server:      server,
				Start:       timeNow(),
				End:         timeNow(),
				TraceID:     traceID,
				ClientName:  clientName,
				ClientAddr:  clientAddr,
				ServerName:  serverName,
				ServerAddr:  serverAddr,
				ClassMethod: classMethod,
				Error:       io.EOF,
			},
		},
	}
	for i, tt := range tests {
		out := &TestOutput{}
		trace := verboseTrace{
			out:         out,
			server:      server,
			traceID:     traceID,
			clientName:  clientName,
			clientAddr:  clientAddr,
			serverName:  serverName,
			serverAddr:  serverAddr,
			classMethod: classMethod,
		}
		trace.Request(tt.args)
		trace.Response(tt.err, tt.reply)
		if got, want := out.request, tt.request; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: request: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: request: got %v", i, got)
		}
		if got, want := out.response, tt.response; !reflect.DeepEqual(got, want) {
			t.Fatalf("%d: response: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: response: got %v", i, got)
		}
	}
}
