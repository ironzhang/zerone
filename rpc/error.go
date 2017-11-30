package rpc

import (
	"fmt"

	"google.golang.org/grpc/codes"

	//"google.golang.org/grpc/codes"
)

type ErrorCode interface {
	Code() codes.Code
}

type ErrorCause interface {
	Cause() error
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

type rpcError struct {
	code  codes.Code
	cause error
}

func (e rpcError) Code() codes.Code {
	return e.code
}

func (e rpcError) Cause() error {
	return e.cause
}

func (e rpcError) Error() string {
	return fmt.Sprintf("%s: %v", e.code.String(), e.cause)
}
