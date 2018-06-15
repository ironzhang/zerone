package rpc

import (
	"encoding/json"
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

type jsonError struct {
	Server string `json:",omitempty"`
	Code   int
	Desc   string
	Cause  string
}

func (e Error) Error() string {
	je := jsonError{
		Server: e.server,
		Code:   int(e.code),
		Desc:   e.code.String(),
		Cause:  e.cause.Error(),
	}
	data, _ := json.Marshal(je)
	return string(data)
}
