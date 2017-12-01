package rpc

import (
	"fmt"

	"github.com/ironzhang/zerone/rpc/codes"
)

type ErrorModule interface {
	Module() string
}

type ErrorCode interface {
	Code() codes.Code
}

type ErrorCause interface {
	Cause() error
}

type rpcError struct {
	module string
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

func NewModuleError(module string, code codes.Code, cause error) error {
	return rpcError{
		module: module,
		code:   code,
		cause:  cause,
	}
}

func ModuleErrorf(module string, code codes.Code, format string, a ...interface{}) error {
	return rpcError{
		module: module,
		code:   code,
		cause:  fmt.Errorf(format, a...),
	}
}

func (e rpcError) Module() string {
	if e.module != "" {
		return e.module
	}
	if me, ok := e.cause.(ErrorModule); ok {
		return me.Module()
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
	if e.module == "" {
		return fmt.Sprintf("code=%d, desc=%s, cause=%v", e.code, e.code.String(), e.cause)
	}
	return fmt.Sprintf("module=%s, code=%d, desc=%s, cause=%v", e.module, e.code, e.code.String(), e.cause)
}
