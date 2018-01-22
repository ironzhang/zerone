package codec

type RequestHeader struct {
	ServiceMethod string // 服务方法名
	Sequence      uint64 // 序号
	TraceID       string // TraceID
	ClientName    string // 客户端名称
	Verbose       int    // 日志详情等级
}

type Error struct {
	Code   int    // 错误码
	Desc   string // 错误描述
	Cause  string // 错误原因
	Module string // 产生错误的模块名称
}

type ResponseHeader struct {
	ServiceMethod string // 服务方法名
	Sequence      uint64 // 序号
	Error         Error  // 错误
}

type ClientCodec interface {
	WriteRequest(*RequestHeader, interface{}) error
	ReadResponseHeader(*ResponseHeader) error
	ReadResponseBody(interface{}) error
	Close() error
}

type ServerCodec interface {
	ReadRequestHeader(*RequestHeader) error
	ReadRequestBody(interface{}) error
	WriteResponse(*ResponseHeader, interface{}) error
	Close() error
}
