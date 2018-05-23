package zclient

import (
	"github.com/ironzhang/zerone/pkg/balance"
	"github.com/ironzhang/zerone/pkg/route"
)

// 负载均衡策略
type BalancePolicy string

// 负载均衡策略常量定义
const (
	RandomBalancer     BalancePolicy = "RandomBalancer"
	RoundRobinBalancer BalancePolicy = "RoundRobinBalancer"
	HashBalancer       BalancePolicy = "HashBalancer"
	NodeBalancer       BalancePolicy = "NodeBalancer"
)

// 负载均衡器集合
type balancerset struct {
	randomBalancer     *balance.RandomBalancer
	roundRobinBalancer *balance.RoundRobinBalancer
	hashBalancer       *balance.HashBalancer
	nodeBalancer       *balance.NodeBalancer
}

func newBalancerset(table route.Table) *balancerset {
	return &balancerset{
		randomBalancer:     balance.NewRandomBalancer(table),
		roundRobinBalancer: balance.NewRoundRobinBalancer(table),
		hashBalancer:       balance.NewHashBalancer(table, nil),
		nodeBalancer:       balance.NewNodeBalancer(table),
	}
}

func (p *balancerset) getLoadBalancer(policy BalancePolicy) balance.LoadBalancer {
	switch policy {
	case RandomBalancer:
		return p.randomBalancer
	case RoundRobinBalancer:
		return p.roundRobinBalancer
	case HashBalancer:
		return p.hashBalancer
	case NodeBalancer:
		return p.nodeBalancer
	default:
		return p.randomBalancer
	}
}
