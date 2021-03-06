package zerone

import (
	"fmt"

	"github.com/ironzhang/x-pearls/govern"
	"github.com/ironzhang/zerone/pkg/route/dtable"
	"github.com/ironzhang/zerone/pkg/route/stable"
	"github.com/ironzhang/zerone/zclient"
	"github.com/ironzhang/zerone/zserver"
)

type SOptions struct {
	Filename string
}

type SZerone struct {
	tables stable.Tables
}

func NewSZerone(opts SOptions) (*SZerone, error) {
	return new(SZerone).Init(opts)
}

func (p *SZerone) Init(opts SOptions) (*SZerone, error) {
	var err error
	if p.tables, err = stable.LoadTables(opts.Filename); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *SZerone) Close() error {
	return nil
}

func (p *SZerone) NewClient(name, service string) (*zclient.Client, error) {
	tb, err := p.tables.Lookup(service)
	if err != nil {
		return nil, err
	}
	return zclient.New(name, tb), nil
}

func (p *SZerone) NewServer(name, service string) (*zserver.Server, error) {
	return zserver.New(name, service, nil), nil
}

type DOptions struct {
	Namespace string
	Driver    string
	Config    interface{}
}

type DZerone struct {
	driver govern.Driver
}

func NewDZerone(opts DOptions) (*DZerone, error) {
	return new(DZerone).Init(opts)
}

func (p *DZerone) Init(opts DOptions) (*DZerone, error) {
	var err error
	if p.driver, err = govern.Open(opts.Driver, opts.Namespace, opts.Config); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *DZerone) Close() error {
	return p.driver.Close()
}

func (p *DZerone) NewClient(name, service string) (*zclient.Client, error) {
	tb := dtable.NewTable(p.driver, service)
	return zclient.New(name, tb), nil
}

func (p *DZerone) NewServer(name, service string) (*zserver.Server, error) {
	return zserver.New(name, service, p.driver), nil
}

type Options interface{}

type Zerone interface {
	NewClient(name, service string) (*zclient.Client, error)
	NewServer(name, service string) (*zserver.Server, error)
	Close() error
}

func NewZerone(opts Options) (Zerone, error) {
	switch o := opts.(type) {
	case SOptions:
		return NewSZerone(o)
	case *SOptions:
		return NewSZerone(*o)
	case DOptions:
		return NewDZerone(o)
	case *DOptions:
		return NewDZerone(*o)
	default:
		return nil, fmt.Errorf("unknown %T options type", opts)
	}
}
