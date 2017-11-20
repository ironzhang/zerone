package codec

type RequestHeader struct {
	Method     string // 方法
	Sequence   uint64 // 序号
	TraceID    string // TraceID
	ClientName string // 客户端名称
	Verbose    bool   // 是否打印日志详情
}

type Error struct {
	Code        int    // 错误码
	Message     string // 错误消息
	Description string // 错误的具体信息
	ServerName  string // 产生错误的服务端名称
}

type ResponseHeader struct {
	Method   string // 方法
	Sequence uint64 // 序号
	Error    Error  // 错误
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
