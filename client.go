package zerone

import "context"

type Client struct {
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) WithLoadBalancer(lb LoadBalancer) *Client {
	return &Client{}
}

func (c *Client) WithFailPolicy(fp FailPolicy) *Client {
	return &Client{}
}

func (c *Client) Call(ctx context.Context, method string, args, res interface{}) error {
	return nil
}
