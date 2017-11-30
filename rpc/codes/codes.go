package codes

import "fmt"

type Code int

const (
	OK Code = 0

	Unknown  Code = -1
	Internal Code = -2

	InvalidHeader  Code = -101
	InvalidRequest Code = -102

	OutOfRange Code = -201
)

var codes = map[Code]string{}

func Register(code Code, desc string) {
	_, ok := codes[code]
	if ok {
		panic(fmt.Sprintf("%d:%s code is registered", code, desc))
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
	Register(InvalidHeader, "invalid header")
	Register(InvalidRequest, "invalid request")
	Register(OutOfRange, "out of range")
}
