package codes_test

import (
	"fmt"
	"testing"

	"github.com/ironzhang/zerone/rpc/codes"
)

func TestRegisterCodes(t *testing.T) {
	tests := []struct {
		code codes.Code
		desc string
	}{
		//{code: -1, desc: "-1"},
		{code: 1, desc: "1"},
		{code: 2, desc: "2"},
		{code: 100, desc: "100"},
		{code: 101, desc: "100"},
	}

	for _, tt := range tests {
		codes.Register(tt.code, tt.desc)
	}
	for _, tt := range tests {
		if got, want := tt.code.String(), tt.desc; got != want {
			t.Errorf("%q != %q", got, want)
		}
	}
}

func TestPrintCodes(t *testing.T) {
	codes := []codes.Code{
		codes.OK,
		codes.Unknown,
		codes.Internal,
		codes.InvalidHeader,
		codes.InvalidRequest,
	}
	for _, code := range codes {
		fmt.Println(code)
	}
}
