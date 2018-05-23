package zserver

import (
	"net"
	"time"

	"github.com/ironzhang/zerone/govern"
	"github.com/ironzhang/zerone/rpc"
)

type Driver interface {
	NewProvider(service string, interval time.Duration, f govern.GetEndpointFunc) govern.Provider
}

type Server struct {
	server   *rpc.Server
	service  string
	driver   Driver
	listener net.Listener
}

func New(name, service string, driver Driver) *Server {
	return &Server{
		server:  rpc.NewServer(name),
		service: service,
		driver:  driver,
	}
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) Register(rcvr interface{}) error {
	return s.server.Register(rcvr)
}

func (s *Server) ListenAndServe(network, address string) (err error) {
	s.listener, err = net.Listen(network, address)
	if err != nil {
		return err
	}
	if s.driver != nil {
		p := s.driver.NewProvider(s.service, 10*time.Second, nil)
		defer p.Close()
	}
	s.server.Accept(s.listener)
	return nil
}
