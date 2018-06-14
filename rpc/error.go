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

type Error struct {
	server string
	code   codes.Code
	cause  error
}

func NewError(code codes.Code, cause error) error {
	return Error{
		code:  code,
		cause: cause,
	}
}

func Errorf(code codes.Code, format string, a ...interface{}) error {
	return Error{
		code:  code,
		cause: fmt.Errorf(format, a...),
	}
}

func NewServerError(server string, code codes.Code, cause error) error {
	return Error{
		server: server,
		code:   code,
		cause:  cause,
	}
}

func ServerErrorf(server string, code codes.Code, format string, a ...interface{}) error {
	return Error{
		server: server,
		code:   code,
		cause:  fmt.Errorf(format, a...),
	}
}

func (e Error) Server() string {
	if e.server != "" {
		return e.server
	}
	if me, ok := e.cause.(ErrorServer); ok {
		return me.Server()
	}
	return ""
}

func (e Error) Code() codes.Code {
	return e.code
}

func (e Error) Cause() error {
	return e.cause
}

func (e Error) Error() string {
	if e.server == "" {
		return fmt.Sprintf("{code: %d, desc: %s, cause: %v}", e.code, e.code.String(), e.cause)
	}
	return fmt.Sprintf("{server: %s, code: %d, desc: %s, cause: %v}", e.server, e.code, e.code.String(), e.cause)
}
