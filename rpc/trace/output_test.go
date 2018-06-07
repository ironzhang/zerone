package trace

import (
	"io"
	"testing"
	"time"
)

func TestStdOutput(t *testing.T) {
	out := NewStdOutput(nil)

	requests := []Request{
		{
			Server:      true,
			Start:       time.Now(),
			TraceID:     "TraceID",
			ClientName:  "ClientName",
			ClientAddr:  "ClientAddr",
			ServerName:  "ServerName",
			ServerAddr:  "ServerAddr",
			ClassMethod: "ClassMethod",
			Args:        nil,
		},
		{
			Server:      false,
			Start:       time.Now(),
			TraceID:     "TraceID",
			ClientName:  "ClientName",
			ClientAddr:  "ClientAddr",
			ServerName:  "ServerName",
			ServerAddr:  "ServerAddr",
			ClassMethod: "ClassMethod",
			Args:        nil,
		},
	}
	for _, r := range requests {
		out.Request(r)
	}

	responses := []Response{
		{
			Server:      true,
			Start:       time.Now(),
			End:         time.Now(),
			TraceID:     "TraceID",
			ClientName:  "ClientName",
			ClientAddr:  "ClientAddr",
			ServerName:  "ServerName",
			ServerAddr:  "ServerAddr",
			ClassMethod: "ClassMethod",
			Reply:       nil,
		},
		{
			Server:      true,
			Start:       time.Now(),
			End:         time.Now(),
			TraceID:     "TraceID",
			ClientName:  "ClientName",
			ClientAddr:  "ClientAddr",
			ServerName:  "ServerName",
			ServerAddr:  "ServerAddr",
			ClassMethod: "ClassMethod",
			Error:       io.EOF,
		},
		{
			Server:      false,
			Start:       time.Now(),
			End:         time.Now(),
			TraceID:     "TraceID",
			ClientName:  "ClientName",
			ClientAddr:  "ClientAddr",
			ServerName:  "ServerName",
			ServerAddr:  "ServerAddr",
			ClassMethod: "ClassMethod",
			Reply:       nil,
		},
		{
			Server:      false,
			Start:       time.Now(),
			End:         time.Now(),
			TraceID:     "TraceID",
			ClientName:  "ClientName",
			ClientAddr:  "ClientAddr",
			ServerName:  "ServerName",
			ServerAddr:  "ServerAddr",
			ClassMethod: "ClassMethod",
			Error:       io.EOF,
		},
	}
	for _, r := range responses {
		out.Response(r)
	}
}
