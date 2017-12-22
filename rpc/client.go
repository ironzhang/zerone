package rpc

import (
	"context"
	"sync"

	"github.com/ironzhang/zerone/rpc/codec"
)

type Call struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
}

type Client struct {
	codec    codec.ClientCodec
	pending  sync.Map
	sequence uint64
}

//func NewClientWithCodec(c codec.ClientCodec) *Client {
//	client := &Client{codec: c}
//	go client.reading()
//	return client
//}

func (c *Client) reading() {
	//	var err error
	//	for {
	//		var resp codec.ResponseHeader
	//		if err = c.codec.ReadResponseHeader(&resp); err != nil {
	//			break
	//		}
	//		c.pending.Load(resp.Sequence)
	//	}
}

func (c *Client) send(call *Call) error {
	return nil
}

func (c *Client) Go(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *Call) error {
	return nil
}

func (c *Client) Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	return nil
}
