package zclient

import (
	"sync"

	"github.com/ironzhang/zerone/rpc"
	"github.com/ironzhang/zerone/rpc/trace"
)

type connector struct {
	name    string
	mu      sync.RWMutex
	output  trace.Output
	verbose int
	clients map[string]*rpc.Client
}

func newConnector(name string) *connector {
	return &connector{
		name:    name,
		output:  trace.DefaultOutput,
		verbose: 0,
		clients: make(map[string]*rpc.Client),
	}
}

func (p *connector) close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.clients {
		c.Close()
	}
	p.clients = make(map[string]*rpc.Client)
}

func (p *connector) setTraceOutput(output trace.Output) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.output = output
	for _, c := range p.clients {
		c.SetTraceOutput(output)
	}
}

func (p *connector) getTraceVerbose() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.verbose
}

func (p *connector) setTraceVerbose(verbose int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.verbose = verbose
	for _, c := range p.clients {
		c.SetTraceVerbose(verbose)
	}
}

func (p *connector) dial(key, net, addr string) (*rpc.Client, error) {
	if c, ok := p.loadClient(key); ok {
		if c.IsShutdown() {
			return nil, rpc.ErrShutdown
		} else if c.IsAvailable() {
			return c, nil
		} else {
			p.deleteClient(key)
			c.Close()
		}
	}

	c, err := rpc.Dial(p.name, net, addr)
	if err != nil {
		return nil, err
	}
	c.SetTraceOutput(p.output)
	c.SetTraceVerbose(p.verbose)

	actual, loaded := p.loadOrStoreClient(key, c)
	if loaded {
		c.Close()
	}
	return actual, nil
}

func (p *connector) loadClient(key string) (*rpc.Client, bool) {
	p.mu.RLock()
	c, ok := p.clients[key]
	p.mu.RUnlock()
	return c, ok
}

func (p *connector) loadOrStoreClient(key string, c *rpc.Client) (actual *rpc.Client, loaded bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if oc, ok := p.clients[key]; ok {
		return oc, true
	}
	p.clients[key] = c
	return c, false
}

func (p *connector) deleteClient(key string) {
	p.mu.Lock()
	delete(p.clients, key)
	p.mu.Unlock()
}
