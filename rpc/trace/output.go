package trace

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Request struct {
	Server      bool
	Start       time.Time
	TraceID     string
	ClientName  string
	ClientAddr  string
	ServerName  string
	ServerAddr  string
	ClassMethod string
	Args        interface{}
}

type Response struct {
	Server      bool
	Start       time.Time
	End         time.Time
	TraceID     string
	ClientName  string
	ClientAddr  string
	ServerName  string
	ServerAddr  string
	ClassMethod string
	Error       error
	Reply       interface{}
}

type Output interface {
	Request(r Request)
	Response(r Response)
}

var DefaultOutput Output = NewStdOutput(nil)

type StdOutput struct {
	w io.Writer
}

func NewStdOutput(w io.Writer) *StdOutput {
	if w == nil {
		w = os.Stdout
	}
	return &StdOutput{w: w}
}

const timeLayout = "2006-01-02 15:04:05.999999 -0700 MST"

func (p *StdOutput) Request(r Request) {
	prefix := "Client"
	if r.Server {
		prefix = "Server"
	}

	args, _ := json.Marshal(r.Args)
	fmt.Fprintf(p.w, "%s %s.Request[%s][%s:%s->%s:%s][%s]: %s\n",
		r.Start.Format(timeLayout),
		prefix,
		r.TraceID,
		r.ClientName,
		r.ClientAddr,
		r.ServerName,
		r.ServerAddr,
		r.ClassMethod,
		args,
	)
}

func (p *StdOutput) Response(r Response) {
	prefix := "Client"
	if r.Server {
		prefix = "Server"
	}

	if r.Error != nil {
		fmt.Fprintf(p.w, "%s %s.Error[%s][%s:%s->%s:%s][%s][%s]: %s\n",
			r.Start.Format(timeLayout),
			prefix,
			r.TraceID,
			r.ClientName,
			r.ClientAddr,
			r.ServerName,
			r.ServerAddr,
			r.ClassMethod,
			r.End.Sub(r.Start),
			r.Error,
		)
	} else {
		reply, _ := json.Marshal(r.Reply)
		fmt.Fprintf(p.w, "%s %s.Reply[%s][%s:%s->%s:%s][%s][%s]: %s\n",
			r.Start.Format(timeLayout),
			prefix,
			r.TraceID,
			r.ClientName,
			r.ClientAddr,
			r.ServerName,
			r.ServerAddr,
			r.ClassMethod,
			r.End.Sub(r.Start),
			reply,
		)
	}
}
