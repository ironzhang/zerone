package rpc

import (
	"fmt"
	"reflect"
	"strings"
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
}

func (s *Server) readRequest(c codec.ServerCodec) (req *codec.RequestHeader, method reflect.Method, rcvr, args, reply reflect.Value, err error) {
	var h codec.RequestHeader
	if err = c.ReadRequestHeader(&h); err != nil {
		return
	}
	req = &h

	serviceName, methodName, err := parseServiceMethod(h.ServiceMethod)
	if err != nil {
		return
	}
	rcvr, meth, err := s.lookupServiceMethod(serviceName, methodName)
	if err != nil {
		return
	}

	args = meth.newArgsValue()
	if err = c.ReadRequestBody(args.Interface()); err != nil {
		return
	}
	reply = meth.newReplyValue()

	return
}

func (s *Server) writeRequest(c codec.ServerCodec) {
}

func (s *Server) lookupServiceMethod(serviceName, methodName string) (reflect.Value, *method, error) {
	svci, ok := s.serviceMap.Load(serviceName)
	if !ok {
		return reflect.Value{}, nil, fmt.Errorf("can't find service %s.%s", serviceName, methodName)
	}
	svc := svci.(*service)
	meth, ok := svc.methods[methodName]
	if !ok {
		return reflect.Value{}, nil, fmt.Errorf("can't find method %s.%s", serviceName, methodName)
	}
	return svc.rcvr, meth, nil
}

func parseServiceMethod(serviceMethod string) (string, string, error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		return "", "", fmt.Errorf("service/method request ill-formed: %s", serviceMethod)
	}
	return serviceMethod[:dot], serviceMethod[dot+1:], nil
}
