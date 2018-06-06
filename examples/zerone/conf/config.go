package conf

import (
	"fmt"

	"github.com/coreos/etcd/client"
	"github.com/ironzhang/x-pearls/config"
	"github.com/ironzhang/x-pearls/govern/etcdv2"
	"github.com/ironzhang/zerone"
)

type ZeroneConfig struct {
	Zerone    string
	Filename  string
	Namespace string
	Endpoints []string
}

func LoadZeroneConfig(filename string) (cfg ZeroneConfig, err error) {
	err = config.LoadFromFile(filename, &cfg)
	return cfg, err
}

func (c ZeroneConfig) ZeroneOptions() (zerone.Options, error) {
	switch c.Zerone {
	case "SZerone":
		return zerone.SOptions{
			Filename: c.Filename,
		}, nil
	case "DZerone":
		return zerone.DOptions{
			Namespace: c.Namespace,
			Driver:    etcdv2.DriverName,
			Config:    client.Config{Endpoints: c.Endpoints},
		}, nil
	default:
		return nil, fmt.Errorf("unknown %q zerone", c.Zerone)
	}
}
