package rpc

import (
	"fmt"
	"reflect"
	"sync"
)

type Server struct {
	mu         sync.RWMutex // protects the serviceMap
	serviceMap map[string]*service
}

func (s *Server) Register(rcvr interface{}) error {
	return s.register(rcvr, "")
}

func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.register(rcvr, name)
}

func (s *Server) register(rcvr interface{}, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.serviceMap == nil {
		s.serviceMap = make(map[string]*service)
	}

	typ := reflect.TypeOf(rcvr)
	val := reflect.ValueOf(rcvr)

	tname := reflect.Indirect(val).Type().Name()
	if !isExported(tname) {
		return fmt.Errorf("register: type %s is not exported", tname)
	}
	if name == "" {
		name = tname
	}
	if name == "" {
		return fmt.Errorf("register: no service name for type %s", typ.Name())
	}
	if _, present := s.serviceMap[name]; present {
		return fmt.Errorf("register: service already defined: %s", name)
	}

	svc, err := parseService(val)
	if err != nil {
		return fmt.Errorf("register: parse service: %v", err)
	}
	if len(svc.methods) <= 0 {
		return fmt.Errorf("register: type %s has no exported methods of suitable type", tname)
	}
	s.serviceMap[name] = svc

	return nil
}
