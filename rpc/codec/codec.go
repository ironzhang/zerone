package codec

type RequestHeader struct {
	Method   string
	Sequence uint64
	TraceID  string
	ClientID string
	Verbose  bool
}

type ResponseHeader struct {
	Method   string
	Sequence uint64
	Code     int
	Desc     string
	Cause    string
}

type ClientCodec interface {
	WriteRequest(*RequestHeader, interface{}) error
	ReadResponseHeader(*ResponseHeader) error
	ReadResponseBody(interface{}) error
}

type ServerCodec interface {
	ReadRequestHeader(*RequestHeader) error
	ReadRequestBody(interface{}) error
	WriteResponse(*ResponseHeader, interface{}) error
}
