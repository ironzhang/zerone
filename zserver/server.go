package zserver

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ironzhang/x-pearls/govern"
	"github.com/ironzhang/zerone/pkg/endpoint"
	"github.com/ironzhang/zerone/rpc"
	"github.com/ironzhang/zerone/rpc/trace"
)

type Server struct {
	server  *rpc.Server
	service string
	driver  govern.Driver

	mu  sync.Mutex
	lns []net.Listener
}

func New(name, service string, driver govern.Driver) *Server {
	return &Server{
		server:  rpc.NewServer(name),
		service: service,
		driver:  driver,
	}
}

func (s *Server) Close() error {
	s.mu.Lock()
	for _, ln := range s.lns {
		ln.Close()
	}
	s.lns = nil
	s.mu.Unlock()
	return nil
}

func (s *Server) SetTraceOutput(out trace.Output) {
	s.server.SetTraceOutput(out)
}

func (s *Server) GetTraceVerbose() int {
	return s.server.GetTraceVerbose()
}

func (s *Server) SetTraceVerbose(verbose int) {
	s.server.SetTraceVerbose(verbose)
}

func (s *Server) Register(rcvr interface{}) error {
	return s.server.Register(rcvr)
}

func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.server.RegisterName(name, rcvr)
}

func (s *Server) ListenAndServe(network, address, endpointName string) (err error) {
	ln, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	s.addListener(ln)

	if s.driver != nil {
		if endpointName == "" {
			endpointName = fmt.Sprintf("%s@%s", network, address)
		}
		p := s.driver.NewProvider(s.service, 10*time.Second, func() govern.Endpoint {
			return &endpoint.Endpoint{
				Name: endpointName,
				Net:  network,
				Addr: address,
			}
		})
		defer p.Close()
	}
	s.server.Accept(ln)
	return nil
}

func (s *Server) addListener(ln net.Listener) {
	s.mu.Lock()
	s.lns = append(s.lns, ln)
	s.mu.Unlock()
}
