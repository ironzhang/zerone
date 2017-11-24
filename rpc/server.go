package rpc

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/ironzhang/zerone/rpc/codec"
)

type Server struct {
	name       string
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

func (s *Server) serveCodec(c codec.ServerCodec) {
	for {
		req, method, rcvr, args, reply, err := s.readRequest(c)
		if err != nil {
			if req != nil {
				s.writeResponse(c, req, reply, err)
			}
		}

		ctx := context.Background()
		rets := method.Func.Call([]reflect.Value{rcvr, reflect.ValueOf(ctx), args, reply})
		erri := rets[0].Interface()
		if erri != nil {
			err = erri.(error)
		}
		s.writeResponse(c, req, reply, err)
	}
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

func (s *Server) writeResponse(c codec.ServerCodec, req *codec.RequestHeader, reply interface{}, err error) error {
	var resp codec.ResponseHeader
	resp.ServiceMethod = req.ServiceMethod
	resp.Sequence = req.Sequence
	if err != nil {
		resp.Error.Code = -1
		resp.Error.Message = "rpc error"
		resp.Error.Description = err.Error()
		resp.Error.ServerName = s.name
	}
	return c.WriteResponse(&resp, reply)
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
