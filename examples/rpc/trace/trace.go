package trace

import (
	"context"

	"github.com/ironzhang/zerone/rpc"
)

type Trace struct {
	s *rpc.Server
}

func New(s *rpc.Server) *Trace {
	return &Trace{s: s}
}

func (p *Trace) SetVerbose(ctx context.Context, verbose int, reply interface{}) error {
	p.s.SetTraceVerbose(verbose)
	return nil
}
