package codes

import "fmt"

type Code int

const (
	OK Code = 0

	Unknown  Code = -1
	Internal Code = -2

	InvalidHeader   Code = -101
	InvalidRequest  Code = -102
	InvalidResponse Code = -103
)

var codes = map[Code]string{}

func Register(code Code, desc string) {
	registered, ok := codes[code]
	if ok {
		panic(fmt.Sprintf("code=%d(%s,%s) is registered", code, desc, registered))
	}
	codes[code] = desc
}

func (c Code) String() string {
	if desc, ok := codes[c]; ok {
		return desc
	}
	return fmt.Sprintf("code(%d)", c)
}

func init() {
	Register(OK, "ok")
	Register(Unknown, "unknown")
	Register(Internal, "internal")
	Register(InvalidHeader, "invalid rpc header")
	Register(InvalidRequest, "invalid rpc request")
	Register(InvalidResponse, "invalid rpc response")
}
