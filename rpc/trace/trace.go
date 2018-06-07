package trace

import (
	"time"
)

var timeNow = time.Now

type Trace interface {
	Request(args interface{})
	Response(err error, reply interface{})
}

type nopTrace struct{}

func (p nopTrace) Request(args interface{}) {
}

func (p nopTrace) Response(err error, reply interface{}) {
}

type errTrace struct {
	out         Output
	server      bool
	traceID     string
	clientName  string
	clientAddr  string
	serverName  string
	serverAddr  string
	classMethod string
	start       time.Time
	args        interface{}
}

func (p *errTrace) Request(args interface{}) {
	p.start = timeNow()
	p.args = args
}

func (p *errTrace) Response(err error, reply interface{}) {
	if err != nil {
		end := timeNow()
		p.out.Request(Request{
			Server:      p.server,
			Start:       p.start,
			TraceID:     p.traceID,
			ClientName:  p.clientName,
			ClientAddr:  p.clientAddr,
			ServerName:  p.serverName,
			ServerAddr:  p.serverAddr,
			ClassMethod: p.classMethod,
			Args:        p.args,
		})
		p.out.Response(Response{
			Server:      p.server,
			Start:       p.start,
			End:         end,
			TraceID:     p.traceID,
			ClientName:  p.clientName,
			ClientAddr:  p.clientAddr,
			ServerName:  p.serverName,
			ServerAddr:  p.serverAddr,
			ClassMethod: p.classMethod,
			Error:       err,
		})
	}
}

type verboseTrace struct {
	out         Output
	server      bool
	traceID     string
	clientName  string
	clientAddr  string
	serverName  string
	serverAddr  string
	classMethod string
	start       time.Time
}

func (p *verboseTrace) Request(args interface{}) {
	p.start = timeNow()
	p.out.Request(Request{
		Server:      p.server,
		Start:       p.start,
		TraceID:     p.traceID,
		ClientName:  p.clientName,
		ClientAddr:  p.clientAddr,
		ServerName:  p.serverName,
		ServerAddr:  p.serverAddr,
		ClassMethod: p.classMethod,
		Args:        args,
	})
}

func (p *verboseTrace) Response(err error, reply interface{}) {
	p.out.Response(Response{
		Server:      p.server,
		Start:       p.start,
		End:         timeNow(),
		TraceID:     p.traceID,
		ClientName:  p.clientName,
		ClientAddr:  p.clientAddr,
		ServerName:  p.serverName,
		ServerAddr:  p.serverAddr,
		ClassMethod: p.classMethod,
		Error:       err,
		Reply:       reply,
	})
}
