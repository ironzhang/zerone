package rpc

import "context"

type Client struct {
}

func (c *Client) Call(ctx context.Context, method string, args interface{}, reply interface{}) error {
	return nil
}
