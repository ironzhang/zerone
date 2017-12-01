package rpc

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/ironzhang/zerone/rpc/codes"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		code  codes.Code
		cause error
	}{
		{code: codes.Code(1), cause: errors.New("code1 error")},
		{code: codes.Code(2), cause: errors.New("code2 error")},
	}
	for _, tt := range tests {
		err := NewError(tt.code, tt.cause)
		emodule := err.(ErrorModule)
		if got, want := emodule.Module(), ""; got != want {
			t.Errorf("module: %q != %q", got, want)
		}
		ecode := err.(ErrorCode)
		if got, want := ecode.Code(), tt.code; got != want {
			t.Errorf("code: %d(%[1]s) != %d(%[2]s)", got, want)
		}
		ecause := err.(ErrorCause)
		if got, want := ecause.Cause(), tt.cause; got != want {
			t.Errorf("cause: %v != %v", got, want)
		}
		t.Log(err)
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		code   codes.Code
		format string
		args   []interface{}
	}{
		{code: codes.Code(1), format: "code(%d) error", args: []interface{}{1}},
		{code: codes.Code(2), format: "code(%d) error", args: []interface{}{2}},
	}
	for _, tt := range tests {
		err := Errorf(tt.code, tt.format, tt.args...)
		emodule := err.(ErrorModule)
		if got, want := emodule.Module(), ""; got != want {
			t.Errorf("module: %q != %q", got, want)
		}
		ecode := err.(ErrorCode)
		if got, want := ecode.Code(), tt.code; got != want {
			t.Errorf("code: %d(%[1]s) != %d(%[2]s)", got, want)
		}
		ecause := err.(ErrorCause)
		if got, want := ecause.Cause().Error(), fmt.Errorf(tt.format, tt.args...).Error(); got != want {
			t.Errorf("cause: %v != %v", got, want)
		}
		t.Log(err)
	}
}

func TestNewModuleError(t *testing.T) {
	tests := []struct {
		module string
		code   codes.Code
		cause  error
	}{
		{module: "module1", code: codes.Code(1), cause: errors.New("code1 error")},
		{module: "module2", code: codes.Code(2), cause: errors.New("code2 error")},
	}
	for _, tt := range tests {
		err := NewModuleError(tt.module, tt.code, tt.cause)
		emodule := err.(ErrorModule)
		if got, want := emodule.Module(), tt.module; got != want {
			t.Errorf("module: %q != %q", got, want)
		}
		ecode := err.(ErrorCode)
		if got, want := ecode.Code(), tt.code; got != want {
			t.Errorf("code: %d(%[1]s) != %d(%[2]s)", got, want)
		}
		ecause := err.(ErrorCause)
		if got, want := ecause.Cause(), tt.cause; got != want {
			t.Errorf("cause: %v != %v", got, want)
		}
		t.Log(err)
	}
}

func TestModuleErrorf(t *testing.T) {
	tests := []struct {
		module string
		code   codes.Code
		format string
		args   []interface{}
	}{
		{module: "module1", code: codes.Code(1), format: "code(%d) error", args: []interface{}{1}},
		{module: "module1", code: codes.Code(2), format: "code(%d) error", args: []interface{}{2}},
	}
	for _, tt := range tests {
		err := ModuleErrorf(tt.module, tt.code, tt.format, tt.args...)
		emodule := err.(ErrorModule)
		if got, want := emodule.Module(), tt.module; got != want {
			t.Errorf("module: %q != %q", got, want)
		}
		ecode := err.(ErrorCode)
		if got, want := ecode.Code(), tt.code; got != want {
			t.Errorf("code: %d(%[1]s) != %d(%[2]s)", got, want)
		}
		ecause := err.(ErrorCause)
		if got, want := ecause.Cause().Error(), fmt.Errorf(tt.format, tt.args...).Error(); got != want {
			t.Errorf("cause: %v != %v", got, want)
		}
		t.Log(err)
	}
}

func getErrorModule(err error) string {
	if e, ok := err.(ErrorModule); ok {
		return e.Module()
	}
	return ""
}

func getErrorCode(err error) codes.Code {
	if e, ok := err.(ErrorCode); ok {
		return e.Code()
	}
	return codes.Unknown
}

func getErrorCause(err error) string {
	if e, ok := err.(ErrorCause); ok {
		return e.Cause().Error()
	}
	return err.Error()
}

func TestErrors(t *testing.T) {
	tests := []struct {
		err    error
		module string
		code   codes.Code
		cause  string
	}{
		{
			err:    NewError(codes.Internal, io.EOF),
			module: "",
			code:   codes.Internal,
			cause:  io.EOF.Error(),
		},
		{
			err:    Errorf(codes.Internal, "read: %v", io.EOF),
			module: "",
			code:   codes.Internal,
			cause:  "read: " + io.EOF.Error(),
		},
		{
			err:    NewModuleError("module1", codes.Internal, io.EOF),
			module: "module1",
			code:   codes.Internal,
			cause:  io.EOF.Error(),
		},
		{
			err:    ModuleErrorf("module1", codes.Internal, "read: %v", io.EOF),
			module: "module1",
			code:   codes.Internal,
			cause:  "read: " + io.EOF.Error(),
		},
		{
			err:    io.EOF,
			module: "",
			code:   codes.Unknown,
			cause:  io.EOF.Error(),
		},
		{
			err:    NewError(codes.InvalidRequest, NewError(codes.Internal, io.EOF)),
			module: "",
			code:   codes.InvalidRequest,
			cause:  NewError(codes.Internal, io.EOF).Error(),
		},
		{
			err:    NewModuleError("module1", codes.InvalidRequest, NewError(codes.Internal, io.EOF)),
			module: "module1",
			code:   codes.InvalidRequest,
			cause:  NewError(codes.Internal, io.EOF).Error(),
		},
		{
			err:    NewModuleError("module1", codes.InvalidRequest, NewModuleError("module2", codes.Internal, io.EOF)),
			module: "module1",
			code:   codes.InvalidRequest,
			cause:  NewModuleError("module2", codes.Internal, io.EOF).Error(),
		},
		{
			err:    NewError(codes.InvalidRequest, NewModuleError("module2", codes.Internal, io.EOF)),
			module: "module2",
			code:   codes.InvalidRequest,
			cause:  NewModuleError("module2", codes.Internal, io.EOF).Error(),
		},
	}
	for _, tt := range tests {
		if got, want := getErrorModule(tt.err), tt.module; got != want {
			t.Errorf("module: %q != %q", got, want)
		}
		if got, want := getErrorCode(tt.err), tt.code; got != want {
			t.Errorf("code: %d(%[1]s) != %d(%[2]s)", got, want)
		}
		if got, want := getErrorCause(tt.err), tt.cause; got != want {
			t.Errorf("module: %q != %q", got, want)
		}
		t.Log(tt.err)
	}
}
