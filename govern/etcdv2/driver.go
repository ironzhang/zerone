package etcdv2

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/zerone/govern"
)

const DriverName = "etcdv2"

type Driver struct {
	namespace string
	client    client.Client
	api       client.KeysAPI
}

func (p *Driver) Name() string {
	return DriverName
}

func (p *Driver) Namespace() string {
	return p.namespace
}

func (p *Driver) NewProvider(service string, interval time.Duration, f govern.GetEndpointFunc) govern.Provider {
	return newProvider(p.api, p.dir(service), interval, f)
}

func (p *Driver) NewConsumer(service string, endpoint govern.Endpoint, f govern.RefreshEndpointsFunc) govern.Consumer {
	return newConsumer(p.api, p.dir(service), endpoint, f)
}

func (p *Driver) Close() error {
	return nil
}

func (p *Driver) dir(service string) string {
	return fmt.Sprintf("/%s/%s", p.namespace, service)
}

func Open(namespace string, config interface{}) (govern.Driver, error) {
	c, err := client.New(config.(client.Config))
	if err != nil {
		return nil, err
	}
	if err = c.Sync(context.Background()); err != nil {
		return nil, err
	}
	return &Driver{
		namespace: namespace,
		client:    c,
		api:       client.NewKeysAPI(c),
	}, nil
}

func init() {
	govern.Register(DriverName, Open)
}
