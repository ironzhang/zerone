package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"

	"github.com/ironzhang/pearls/uuid"
	"github.com/ironzhang/zerone/rpc/codec"
	"github.com/ironzhang/zerone/rpc/codec/json_codec"
	"github.com/ironzhang/zerone/rpc/codes"
	"github.com/ironzhang/zerone/zlog"
)

var (
	ErrShutdown    = errors.New("connection is shutdown")
	ErrUnavailable = errors.New("connection is unavailable")
)

type Call struct {
	Header codec.RequestHeader
	Args   interface{}
	Reply  interface{}
	Error  error
	Done   chan *Call

	trace trace
}

func (c *Call) done() {
	if c.trace != nil {
		c.trace.PrintResponse(c.Error, c.Reply)
	}
	select {
	case c.Done <- c:
		// ok
	default:
		zlog.Warn("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

func (c *Call) send(codec codec.ClientCodec) error {
	if err := codec.WriteRequest(&c.Header, c.Args); err != nil {
		return err
	}
	if c.trace != nil {
		c.trace.PrintRequest(c.Args)
	}
	return nil
}

type Client struct {
	name   string
	codec  codec.ClientCodec
	logger traceLogger

	pending   sync.Map
	sequence  uint64
	shutdown  int32
	available int32
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
		logger: traceLogger{out: os.Stdout, verbose: 0},
	}
	go client.reading()
	return client
}

func (c *Client) SetTraceOutput(out io.Writer) {
	c.logger.SetOutput(out)
}

func (c *Client) SetTraceVerbose(verbose int) {
	c.logger.SetVerbose(verbose)
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
		call.Error = ModuleErrorf(resp.Error.Module, codes.Code(resp.Error.Code), resp.Error.Cause)
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
			zlog.Tracef("read response: %v", err)
		}
	}

	atomic.StoreInt32(&c.available, 1)
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

	zlog.Tracef("client quit reading: %v", err)
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

func (c *Client) Go(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *Call) (*Call, error) {
	if atomic.LoadInt32(&c.shutdown) == 1 {
		return nil, ErrShutdown
	}
	if atomic.LoadInt32(&c.available) == 1 {
		return nil, ErrUnavailable
	}

	if done == nil {
		done = make(chan *Call, 10)
	} else {
		if cap(done) == 0 {
			zlog.Panic("rpc: done channel is unbuffered")
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
			ServiceMethod: serviceMethod,
			Sequence:      sequence,
			ClientName:    c.name,
			TraceID:       traceID,
			Verbose:       verbose,
		},
		Args:  args,
		Reply: reply,
		Done:  done,
		trace: c.logger.NewTrace("Client", verbose, traceID, c.name, serviceMethod),
	}
	if err := c.send(call); err != nil {
		return nil, err
	}
	return call, nil
}

func (c *Client) Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	call, err := c.Go(ctx, serviceMethod, args, reply, make(chan *Call, 1))
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
