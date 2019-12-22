package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ironzhang/pearls/uuid"
	log "github.com/ironzhang/tlog"
	"github.com/ironzhang/zerone/rpc/codec"
	"github.com/ironzhang/zerone/rpc/codec/json_codec"
	"github.com/ironzhang/zerone/rpc/codes"
	"github.com/ironzhang/zerone/rpc/trace"
)

var (
	ErrTimeout     = errors.New("remote process call timeout")
	ErrShutdown    = errors.New("connection is shutdown")
	ErrUnavailable = errors.New("connection is unavailable")
)

type Call struct {
	Header codec.RequestHeader
	Args   interface{}
	Reply  interface{}
	Error  error
	Done   chan *Call

	trace trace.Trace
}

func (c *Call) done() {
	if c.trace != nil {
		c.trace.Response(c.Error, c.Reply)
	}
	select {
	case c.Done <- c:
		// ok
	default:
		log.Warn("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

func (c *Call) send(codec codec.ClientCodec) error {
	if err := codec.WriteRequest(&c.Header, c.Args); err != nil {
		return err
	}
	if c.trace != nil {
		c.trace.Request(c.Args)
	}
	return nil
}

type Client struct {
	name   string
	codec  codec.ClientCodec
	logger *trace.Logger

	pending     sync.Map
	sequence    uint64
	shutdown    int32
	unavailable int32
}

func Dial(name, network, address string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(name, conn), nil
}

func NewClient(name string, rwc io.ReadWriteCloser) *Client {
	return NewClientWithCodec(name, json_codec.NewClientCodec(rwc))
}

func NewClientWithCodec(name string, c codec.ClientCodec) *Client {
	client := &Client{
		name:   name,
		codec:  c,
		logger: trace.NewLogger(),
	}
	go client.reading()
	return client
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) SetTraceOutput(out trace.Output) {
	c.logger.SetOutput(out)
}

func (c *Client) GetTraceVerbose() int {
	return c.logger.GetVerbose()
}

func (c *Client) SetTraceVerbose(verbose int) {
	c.logger.SetVerbose(verbose)
}

func (c *Client) IsShutdown() bool {
	return atomic.LoadInt32(&c.shutdown) == 1
}

func (c *Client) IsAvailable() bool {
	return atomic.LoadInt32(&c.unavailable) == 0
}

func (c *Client) readResponse() (keepReading bool, err error) {
	var resp codec.ResponseHeader
	if err = c.codec.ReadResponseHeader(&resp); err != nil {
		return false, err
	}

	value, ok := c.pending.Load(resp.Sequence)
	if !ok {
		c.codec.ReadResponseBody(nil)
		return true, fmt.Errorf("sequence(%d) not found", resp.Sequence)
	}
	c.pending.Delete(resp.Sequence)
	call := value.(*Call)

	if resp.Error.Code != 0 {
		err = c.codec.ReadResponseBody(nil)
		call.Error = ServerErrorf(resp.Error.ServerName, codes.Code(resp.Error.Code), resp.Error.Cause)
		call.done()
		return true, err
	}
	if err = c.codec.ReadResponseBody(call.Reply); err != nil {
		call.Error = NewError(codes.InvalidResponse, err)
	}
	call.done()
	return true, err
}

func (c *Client) reading() {
	var err error
	for keepReading := true; keepReading; {
		if keepReading, err = c.readResponse(); err != nil {
			log.Debugf("read response: %v", err)
		}
	}

	atomic.StoreInt32(&c.unavailable, 1)
	if atomic.LoadInt32(&c.shutdown) == 1 {
		err = ErrShutdown
	} else {
		err = ErrUnavailable
	}
	c.pending.Range(func(key, value interface{}) bool {
		call := value.(*Call)
		call.Error = err
		call.done()
		return true
	})

	log.Debugf("client quit reading: %v", err)
}

func (c *Client) send(call *Call) (err error) {
	if _, loaded := c.pending.LoadOrStore(call.Header.Sequence, call); loaded {
		return fmt.Errorf("sequence(%d) duplicate", call.Header.Sequence)
	}
	if err = call.send(c.codec); err != nil {
		c.pending.Delete(call.Header.Sequence)
		return err
	}
	return nil
}

func (c *Client) Go(ctx context.Context, classMethod string, args interface{}, reply interface{}, timeout time.Duration, done chan *Call) (*Call, error) {
	if c.IsShutdown() {
		return nil, ErrShutdown
	}
	if !c.IsAvailable() {
		return nil, ErrUnavailable
	}

	if done == nil {
		done = make(chan *Call, 10)
	} else {
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}

	sequence := atomic.AddUint64(&c.sequence, 1)
	verbose, _ := ParseVerbose(ctx)
	traceID, ok := ParseTraceID(ctx)
	if !ok {
		traceID = uuid.New().String()
	}

	call := &Call{
		Header: codec.RequestHeader{
			ClassMethod: classMethod,
			Sequence:    sequence,
			ClientName:  c.name,
			TraceID:     traceID,
			Verbose:     verbose,
		},
		Args:  args,
		Reply: reply,
		Done:  done,
		trace: c.logger.NewTrace(false, verbose, traceID, c.name, "", "", "", classMethod),
	}
	if err := c.send(call); err != nil {
		return nil, err
	}

	// 超时处理
	if timeout > 0 {
		time.AfterFunc(timeout, func() {
			if value, ok := c.pending.Load(sequence); ok {
				c.pending.Delete(sequence)
				call := value.(*Call)
				call.Error = ErrTimeout
				call.done()
			}
		})
	}

	return call, nil
}

func (c *Client) Call(ctx context.Context, classMethod string, args interface{}, reply interface{}, timeout time.Duration) error {
	call, err := c.Go(ctx, classMethod, args, reply, timeout, make(chan *Call, 1))
	if err != nil {
		return err
	}
	<-call.Done
	return call.Error
}

func (c *Client) Close() error {
	if atomic.CompareAndSwapInt32(&c.shutdown, 0, 1) {
		return c.codec.Close()
	}
	return ErrShutdown
}
