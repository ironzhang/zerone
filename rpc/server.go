package rpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/ironzhang/zerone/rpc/codec"
	"github.com/ironzhang/zerone/rpc/codec/json_codec"
	"github.com/ironzhang/zerone/rpc/codes"
	"github.com/ironzhang/zerone/zlog"
)

type Server struct {
	name       string
	logger     traceLogger
	serviceMap sync.Map
}

func NewServer(name string) *Server {
	return &Server{
		name:   name,
		logger: traceLogger{out: os.Stdout, verbose: 0},
	}
}

func (s *Server) SetTraceOutput(out io.Writer) {
	s.logger.SetOutput(out)
}

func (s *Server) GetTraceVerbose() int {
	return s.logger.GetVerbose()
}

func (s *Server) SetTraceVerbose(verbose int) {
	s.logger.SetVerbose(verbose)
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

func (s *Server) Register(rcvr interface{}) error {
	return s.register(rcvr, "")
}

func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.register(rcvr, name)
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

func (s *Server) readRequest(c codec.ServerCodec) (req *codec.RequestHeader, method reflect.Method, rcvr, args, reply reflect.Value, keepReading bool, err error) {
	var h codec.RequestHeader
	if err = c.ReadRequestHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			keepReading = true
		}
		return
	}
	req = &h
	keepReading = true

	serviceName, methodName, err := splitServiceMethod(req.ServiceMethod)
	if err != nil {
		err = NewError(codes.InvalidHeader, err)
		c.ReadRequestBody(nil)
		return
	}
	rcvr, meth, err := s.lookupServiceMethod(serviceName, methodName)
	if err != nil {
		err = NewError(codes.InvalidHeader, err)
		c.ReadRequestBody(nil)
		return
	}
	method = meth.method

	argIsValue := false
	if meth.args.Kind() == reflect.Ptr {
		args = reflect.New(meth.args.Elem())
	} else {
		args = reflect.New(meth.args)
		argIsValue = true
	}
	if err = c.ReadRequestBody(args.Interface()); err != nil {
		err = NewError(codes.InvalidRequest, err)
		return
	}
	if argIsValue {
		args = args.Elem()
	}

	if isNilInterface(meth.reply) {
		reply = reflect.New(meth.reply)
	} else {
		reply = reflect.New(meth.reply.Elem())
		switch meth.reply.Elem().Kind() {
		case reflect.Map:
			reply.Elem().Set(reflect.MakeMap(meth.reply.Elem()))
		case reflect.Slice:
			reply.Elem().Set(reflect.MakeSlice(meth.reply.Elem(), 0, 0))
		}
	}

	return
}

func (s *Server) writeResponse(c codec.ServerCodec, req *codec.RequestHeader, reply interface{}, err error) error {
	var resp codec.ResponseHeader
	resp.ServiceMethod = req.ServiceMethod
	resp.Sequence = req.Sequence
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
		module := s.name
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

func (s *Server) call(method reflect.Method, rcvr, args, reply reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	ctx := context.Background()
	rets := method.Func.Call([]reflect.Value{rcvr, reflect.ValueOf(ctx), args, reply})
	erri := rets[0].Interface()
	if erri != nil {
		err = erri.(error)
	}
	return err
}

func (s *Server) rpcError(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(rpcError); ok {
		if e.module == "" {
			e.module = s.name
		}
		return e
	}

	module := s.name
	if e, ok := err.(ErrorModule); ok {
		if m := e.Module(); m != "" {
			module = m
		}
	}
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
	return ModuleErrorf(module, code, cause)
}

var emptyResp = struct{}{}

func (s *Server) serveError(c codec.ServerCodec, req *codec.RequestHeader, err error) {
	tr := s.logger.NewTrace("Server", req.Verbose, req.TraceID, req.ClientName, req.ServiceMethod)
	tr.PrintRequest(nil)
	s.writeResponse(c, req, emptyResp, err)
	tr.PrintResponse(s.rpcError(err), emptyResp)
}

func (s *Server) serveCall(c codec.ServerCodec, req *codec.RequestHeader, method reflect.Method, rcvr, args, reply reflect.Value) {
	tr := s.logger.NewTrace("Server", req.Verbose, req.TraceID, req.ClientName, req.ServiceMethod)
	tr.PrintRequest(args.Interface())
	err := s.call(method, rcvr, args, reply)
	s.writeResponse(c, req, reply.Interface(), err)
	tr.PrintResponse(s.rpcError(err), reply.Interface())
}

func (s *Server) ServeRequest(c codec.ServerCodec) error {
	req, method, rcvr, args, reply, keepReading, err := s.readRequest(c)
	if err != nil {
		if !keepReading {
			return err
		}
		if req != nil {
			s.serveError(c, req, err)
		}
		return err
	}
	s.serveCall(c, req, method, rcvr, args, reply)
	return nil
}

func (s *Server) ServeCodec(c codec.ServerCodec) {
	defer c.Close()
	for {
		req, method, rcvr, args, reply, keepReading, err := s.readRequest(c)
		if err != nil {
			if !keepReading {
				break
			}
			if req != nil {
				s.serveError(c, req, err)
			}
			continue
		}
		go s.serveCall(c, req, method, rcvr, args, reply)
	}
	zlog.Trace("server quit serve codec")
}

func (s *Server) ServeConn(rwc io.ReadWriteCloser) {
	s.ServeCodec(json_codec.NewServerCodec(rwc))
}

func (s *Server) Accept(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			zlog.Tracef("rpc.Accept: %v", err)
			return
		}
		go s.ServeConn(conn)
	}
}
