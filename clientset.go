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

func (p *clientset) add(key, net, addr string) (*rpc.Client, error) {
	p.mu.RLock()
	if c, ok := p.clients[key]; ok {
		p.mu.RUnlock()
		return c, nil
	}
	p.mu.RUnlock()

	c, err := rpc.Dial(p.name, net, addr)
	if err != nil {
		return nil, err
	}
	c.SetTraceOutput(p.output)
	c.SetTraceVerbose(p.verbose)

	p.mu.Lock()
	p.clients[key] = c
	p.mu.Unlock()
	return c, nil
}

func (p *clientset) remove(key string) {
	p.mu.Lock()
	c, ok := p.clients[key]
	if ok {
		delete(p.clients, key)
	}
	p.mu.Unlock()

	if ok {
		c.Close()
	}
}
