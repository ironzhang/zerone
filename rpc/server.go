package rpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"

	"github.com/ironzhang/x-pearls/log"
	"github.com/ironzhang/zerone/rpc/codec"
	"github.com/ironzhang/zerone/rpc/codec/json_codec"
	"github.com/ironzhang/zerone/rpc/codes"
	"github.com/ironzhang/zerone/rpc/trace"
)

type Server struct {
	name     string
	logger   *trace.Logger
	classMap sync.Map
}

func NewServer(name string) *Server {
	return &Server{
		name:   name,
		logger: trace.NewLogger(),
	}
}

func (s *Server) SetTraceOutput(out trace.Output) {
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
		return fmt.Errorf("register: no class name for type %s", typ.Name())
	}
	c, err := parseClass(name, val)
	if err != nil {
		return fmt.Errorf("register: parse class: %v", err)
	}

	if _, loaded := s.classMap.LoadOrStore(name, c); loaded {
		return fmt.Errorf("register: class already defined: %s", name)
	}
	return nil
}

func (s *Server) Register(rcvr interface{}) error {
	return s.register(rcvr, "")
}

func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.register(rcvr, name)
}

func splitClassMethod(classMethod string) (string, string, error) {
	dot := strings.LastIndex(classMethod, ".")
	if dot < 0 {
		return "", "", fmt.Errorf("class/method request ill-formed: %s", classMethod)
	}
	return classMethod[:dot], classMethod[dot+1:], nil
}

func (s *Server) lookupClassMethod(className, methodName string) (reflect.Value, *method, error) {
	v, ok := s.classMap.Load(className)
	if !ok {
		return reflect.Value{}, nil, fmt.Errorf("can't find class %s.%s", className, methodName)
	}
	c := v.(*class)
	meth, ok := c.methods[methodName]
	if !ok {
		return reflect.Value{}, nil, fmt.Errorf("can't find method %s.%s", className, methodName)
	}
	return c.rcvr, meth, nil
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

	className, methodName, err := splitClassMethod(req.ClassMethod)
	if err != nil {
		err = NewError(codes.InvalidHeader, err)
		c.ReadRequestBody(nil)
		return
	}
	rcvr, meth, err := s.lookupClassMethod(className, methodName)
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
	if !isNilInterface(meth.args) {
		if err = c.ReadRequestBody(args.Interface()); err != nil {
			err = NewError(codes.InvalidRequest, err)
			return
		}
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
	resp.ClassMethod = req.ClassMethod
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
		name := s.name
		if e, ok := err.(ErrorServer); ok {
			if sn := e.Server(); sn != "" {
				name = sn
			}
		}

		resp.Error.Code = int(code)
		resp.Error.Desc = code.String()
		resp.Error.Cause = cause
		resp.Error.ServerName = name
	}
	return c.WriteResponse(&resp, reply)
}

func (s *Server) call(req *codec.RequestHeader, method reflect.Method, rcvr, args, reply reflect.Value) (err error) {
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
	ctx = WithTraceID(ctx, req.TraceID)
	ctx = WithVerbose(ctx, req.Verbose)
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
		if e.server == "" {
			e.server = s.name
		}
		return e
	}

	module := s.name
	if e, ok := err.(ErrorServer); ok {
		if m := e.Server(); m != "" {
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
	return ServerErrorf(module, code, cause)
}

var emptyResp = struct{}{}

func (s *Server) serveError(c codec.ServerCodec, req *codec.RequestHeader, err error) {
	tr := s.logger.NewTrace(true, req.Verbose, req.TraceID, req.ClientName, "", s.name, "", req.ClassMethod)
	tr.Request(nil)
	s.writeResponse(c, req, emptyResp, err)
	tr.Response(s.rpcError(err), emptyResp)
}

func (s *Server) serveCall(c codec.ServerCodec, req *codec.RequestHeader, method reflect.Method, rcvr, args, reply reflect.Value) {
	tr := s.logger.NewTrace(true, req.Verbose, req.TraceID, req.ClientName, "", s.name, "", req.ClassMethod)
	tr.Request(args.Interface())
	err := s.call(req, method, rcvr, args, reply)
	s.writeResponse(c, req, reply.Interface(), err)
	tr.Response(s.rpcError(err), reply.Interface())
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
	log.Trace("server quit serve codec")
}

func (s *Server) ServeConn(rwc io.ReadWriteCloser) {
	s.ServeCodec(json_codec.NewServerCodec(rwc))
}

func (s *Server) Accept(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Tracef("rpc.Accept: %v", err)
			return
		}
		go s.ServeConn(conn)
	}
}
