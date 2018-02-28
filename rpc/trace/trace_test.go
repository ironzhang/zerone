package trace

import (
	"io"
	"os"
	"testing"
)

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

func TestErrorTrace(t *testing.T) {
	tr := errorTrace{
		out:           os.Stdout,
		traceID:       "381ec868-369c-41b3-a61a-a1a1b4141c5a",
		clientName:    "client",
		serviceMethod: "server.method",
	}
	tr.PrintRequest(Args{1, 2})
	tr.PrintResponse(nil, Reply{3})
	tr.PrintResponse(io.EOF, nil)
}

func TestVerboseTrace(t *testing.T) {
	tr := verboseTrace{
		out:           os.Stdout,
		traceID:       "6af3b859-9003-4760-a24a-fe4ab16013f0",
		clientName:    "client",
		serviceMethod: "server.method",
	}
	tr.PrintRequest(Args{1, 2})
	tr.PrintResponse(nil, Reply{3})
	tr.PrintResponse(io.EOF, nil)
}
