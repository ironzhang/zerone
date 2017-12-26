package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"github.com/ironzhang/zerone/rpc/codec"
	"github.com/ironzhang/zerone/rpc/codec/json-codec"
	"github.com/ironzhang/zerone/rpc/codes"
)

var (
	ErrShutdown    = errors.New("connection is shutdown")
	ErrUnavailable = errors.New("connection is unavailable")
)

type Call struct {
	Context       context.Context
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

func (c *Call) done() {
	select {
	case c.Done <- c:
		// ok
	default:
		log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

type Client struct {
	name string

	mutex   sync.Mutex
	request codec.RequestHeader
	codec   codec.ClientCodec

	pending   sync.Map
	sequence  uint64
	shutdown  int32
	available int32
}

func NewClient(rwc io.ReadWriteCloser) *Client {
	return NewClientWithCodec(json_codec.NewClientCodec(rwc))
}

func NewClientWithCodec(c codec.ClientCodec) *Client {
	client := &Client{codec: c}
	go client.reading()
	return client
}

func (c *Client) readResponse() (err error) {
	var resp codec.ResponseHeader
	if err = c.codec.ReadResponseHeader(&resp); err != nil {
		return err
	}

	value, ok := c.pending.Load(resp.Sequence)
	if !ok {
		c.codec.ReadResponseBody(nil)
		return fmt.Errorf("sequence(%d) not found", resp.Sequence)
	}
	c.pending.Delete(resp.Sequence)
	call := value.(*Call)

	if resp.Error.Code != 0 {
		call.Error = ModuleErrorf(resp.Error.Module, codes.Code(resp.Error.Code), resp.Error.Cause)
		call.done()
		return c.codec.ReadResponseBody(nil)
	}

	if err = c.codec.ReadResponseBody(call.Reply); err != nil {
		call.Error = NewError(codes.InvalidResponse, err)
	}
	call.done()
	return err
}

func (c *Client) reading() {
	var err error
	for {
		if err = c.readResponse(); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			log.Printf("read response: %v", err)
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
}

func (c *Client) send(call *Call) (err error) {
	if atomic.LoadInt32(&c.shutdown) == 1 {
		return ErrShutdown
	}
	if atomic.LoadInt32(&c.available) == 1 {
		return ErrUnavailable
	}

	sequence := atomic.AddUint64(&c.sequence, 1)
	if _, loaded := c.pending.LoadOrStore(sequence, call); loaded {
		return fmt.Errorf("sequence(%d) duplicate", sequence)
	}

	traceID, ok := ParseTraceID(call.Context)
	if !ok {
		traceID = "new-trace-id"
	}
	verbose, _ := ParseVerbose(call.Context)

	// ClientCodec.WriteRequest不能保证并发安全, 所以这里需要加锁保护
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.request.ServiceMethod = call.ServiceMethod
	c.request.Sequence = sequence
	c.request.TraceID = traceID
	c.request.ClientName = c.name
	c.request.Verbose = verbose
	if err = c.codec.WriteRequest(&c.request, call.Args); err != nil {
		c.pending.Delete(sequence)
		return err
	}
	return nil
}

func (c *Client) Go(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *Call) (*Call, error) {
	if done == nil {
		done = make(chan *Call, 10)
	} else {
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}

	call := &Call{
		Context:       ctx,
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
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
