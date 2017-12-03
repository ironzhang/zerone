package rpc

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/ironzhang/zerone/rpc/codec"
	"github.com/ironzhang/zerone/rpc/codes"
)

type serverRequest struct {
	serviceMethod string
	serviceName   string
	methodName    string
	sequence      uint64
}

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

func splitServiceMethod(serviceMethod string) (string, string, error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		return "", "", fmt.Errorf("service/method request ill-formed: %s", serviceMethod)
	}
	return serviceMethod[:dot], serviceMethod[dot+1:], nil
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

func (s *Server) readRequest(c codec.ServerCodec) (response bool, req serverRequest, method reflect.Method, rcvr, args, reply reflect.Value, err error) {
	var h codec.RequestHeader
	if err = c.ReadRequestHeader(&h); err != nil {
		err = NewError(codes.InvalidHeader, err)
		return
	}

	response = true
	req.serviceMethod = h.ServiceMethod
	req.sequence = h.Sequence

	req.serviceName, req.methodName, err = splitServiceMethod(req.serviceMethod)
	if err != nil {
		err = NewError(codes.InvalidHeader, err)
		c.ReadRequestBody(nil)
		return
	}
	rcvr, meth, err := s.lookupServiceMethod(req.serviceName, req.methodName)
	if err != nil {
		err = NewError(codes.InvalidHeader, err)
		c.ReadRequestBody(nil)
		return
	}

	method = meth.method
	args = meth.newArgsValue()
	if err = c.ReadRequestBody(args.Interface()); err != nil {
		err = NewError(codes.InvalidRequest, err)
		return
	}
	reply = meth.newReplyValue()

	return
}

func (s *Server) writeResponse(c codec.ServerCodec, req serverRequest, reply interface{}, err error) error {
	var resp codec.ResponseHeader
	resp.ServiceMethod = req.serviceMethod
	resp.Sequence = req.sequence
	if err != nil {
		code := codes.Unknown
		if e, ok := err.(ErrorCode); ok {
			code = e.Code()
		}
		cause := err.Error()
		if e, ok := err.(ErrorCause); ok {
			if ce := e.Cause(); ce != nil {
				cause = ce.Error()
			}
		}
		module := req.serviceName
		if e, ok := err.(ErrorModule); ok {
			if m := e.Module(); m != "" {
				module = m
			}
		}

		resp.Error.Code = int(code)
		resp.Error.Desc = code.String()
		resp.Error.Cause = cause
		resp.Error.Module = module
	}
	return c.WriteResponse(&resp, reply)
}

func (s *Server) serveCodec(c codec.ServerCodec) {
	for {
		response, req, method, rcvr, args, reply, err := s.readRequest(c)
		if err != nil {
			if response {
				s.writeResponse(c, req, reply, err)
			}
		}

		ctx := context.Background()
		rets := method.Func.Call([]reflect.Value{rcvr, reflect.ValueOf(ctx), args, reply})
		erri := rets[0].Interface()
		if erri != nil {
			err = erri.(error)
		}
		s.writeResponse(c, req, reply.Interface(), err)
	}
}
