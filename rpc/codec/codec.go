package codec

type RequestHeader struct {
	ClassMethod string // 类方法名, 格式: class.method
	Sequence    uint64 // 序号
	ClientName  string // 客户端名称
	TraceID     string // TraceID
	Verbose     int    // 日志详情等级
}

type Error struct {
	Code       int    // 错误码
	Desc       string // 错误描述
	Cause      string // 错误原因
	ServerName string // 产生错误的服务器名称
}

type ResponseHeader struct {
	ClassMethod string // 类方法名, 格式: class.method
	Sequence    uint64 // 序号
	Error       Error  // 错误
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
