package rpc

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/ironzhang/zerone/rpc/codec"
)

type Server struct {
	serviceMap sync.Map
}

func (s *Server) Register(rcvr interface{}) error {
	return s.register(rcvr, "")
}

func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.register(rcvr, name)
}

func (s *Server) register(rcvr interface{}, name string) error {
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
	svc, err := parseService(name, val)
	if err != nil {
		return fmt.Errorf("register: parse service: %v", err)
	}

	if _, loaded := s.serviceMap.LoadOrStore(name, svc); loaded {
		return fmt.Errorf("register: service already defined: %s", name)
	}
	return nil
}

func (s *Server) serveCodec(codec codec.ServerCodec) {
	for {
	}
}

func (s *Server) readRequest(c codec.ServerCodec) (err error) {
	//	var h codec.RequestHeader
	//	if err = c.ReadRequestHeader(&h); err != nil {
	//		return err
	//	}
	//
	//	dot := strings.LastIndex(h.Method, ".")
	//	if dot < 0 {
	//		return fmt.Errorf("rpc: service/method request ill-formed: %s", h.Method)
	//	}
	//	serviceName := h.Method[:dot]
	//	methodName := h.Method[dot+1:]
	//
	//	svci, ok := s.serviceMap.Load(serviceName)
	//	if !ok {
	//		return fmt.Errorf("rpc: can't find service %s", h.Method)
	//	}
	//	svc := svci.(*service)
	//	m, ok := svc.methods[methodName]
	//	if !ok {
	//		return fmt.Errorf("rpc: can't find method: %s", h.Method)
	//	}
	//
	//	var args reflect.Value
	//	if m.args.Kind() == reflect.Ptr {
	//		args = reflect.New(m.args.Elem())
	//	} else {
	//		args = reflect.New(m.args)
	//	}
	//	if err = c.ReadRequestBody(args.Interface()); err != nil {
	//		return err
	//	}

	return nil
}
