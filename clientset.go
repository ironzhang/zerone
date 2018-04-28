package zerone

import (
	"io"
	"sync"

	"github.com/ironzhang/zerone/rpc"
)

type clientset struct {
	name    string
	mu      sync.RWMutex
	output  io.Writer
	verbose int
	clients map[string]*rpc.Client
}

func newClientset(name string, output io.Writer, verbose int) *clientset {
	return &clientset{
		name:    name,
		output:  output,
		verbose: verbose,
		clients: make(map[string]*rpc.Client),
	}
}

func (p *clientset) close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.clients {
		c.Close()
	}
	p.clients = make(map[string]*rpc.Client)
}

func (p *clientset) setTraceOutput(output io.Writer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.output = output
	for _, c := range p.clients {
		c.SetTraceOutput(output)
	}
}

func (p *clientset) setTraceVerbose(verbose int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.verbose = verbose
	for _, c := range p.clients {
		c.SetTraceVerbose(verbose)
	}
}

func (p *clientset) dial(key, net, addr string) (*rpc.Client, error) {
	p.mu.RLock()
	c, ok := p.clients[key]
	p.mu.RUnlock()
	if ok {
		if c.IsShutdown() {
			return nil, rpc.ErrShutdown
		} else if c.IsAvailable() {
			return c, nil
		} else {
			p.mu.Lock()
			delete(p.clients, key)
			p.mu.Unlock()
			c.Close()
		}
	}

	nc, err := rpc.Dial(p.name, net, addr)
	if err != nil {
		return nil, err
	}
	nc.SetTraceOutput(p.output)
	nc.SetTraceVerbose(p.verbose)

	p.mu.Lock()
	if c, ok = p.clients[key]; ok {
		nc.Close()
		nc = c
	} else {
		p.clients[key] = nc
	}
	p.mu.Unlock()
	return nc, nil
}
