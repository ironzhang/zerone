package rpc

import (
	"fmt"

	"github.com/ironzhang/zerone/rpc/codes"
)

type ErrorServer interface {
	Server() string
}

type ErrorCode interface {
	Code() codes.Code
}

type ErrorCause interface {
	Cause() error
}

type rpcError struct {
	server string
	code   codes.Code
	cause  error
}

func NewError(code codes.Code, cause error) error {
	return rpcError{
		code:  code,
		cause: cause,
	}
}

func Errorf(code codes.Code, format string, a ...interface{}) error {
	return rpcError{
		code:  code,
		cause: fmt.Errorf(format, a...),
	}
}

func NewServerError(server string, code codes.Code, cause error) error {
	return rpcError{
		server: server,
		code:   code,
		cause:  cause,
	}
}

func ServerErrorf(server string, code codes.Code, format string, a ...interface{}) error {
	return rpcError{
		server: server,
		code:   code,
		cause:  fmt.Errorf(format, a...),
	}
}

func (e rpcError) Server() string {
	if e.server != "" {
		return e.server
	}
	if me, ok := e.cause.(ErrorServer); ok {
		return me.Server()
	}
	return ""
}

func (e rpcError) Code() codes.Code {
	return e.code
}

func (e rpcError) Cause() error {
	return e.cause
}

func (e rpcError) Error() string {
	if e.server == "" {
		return fmt.Sprintf("{code: %d, desc: %s, cause: %v}", e.code, e.code.String(), e.cause)
	}
	return fmt.Sprintf("{server: %s, code: %d, desc: %s, cause: %v}", e.server, e.code, e.code.String(), e.cause)
}
