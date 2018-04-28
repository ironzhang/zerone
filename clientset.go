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

func (p *clientset) loadClient(key string) (*rpc.Client, bool) {
	p.mu.RLock()
	c, ok := p.clients[key]
	p.mu.RUnlock()
	return c, ok
}

func (p *clientset) loadOrStoreClient(key string, c *rpc.Client) (actual *rpc.Client, loaded bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if oc, ok := p.clients[key]; ok {
		return oc, true
	}
	p.clients[key] = c
	return c, false
}

func (p *clientset) deleteClient(key string) {
	p.mu.Lock()
	delete(p.clients, key)
	p.mu.Unlock()
}
