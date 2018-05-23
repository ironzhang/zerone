package zclient

import (
	"reflect"
	"testing"

	"github.com/ironzhang/zerone/pkg/balance"
)

func TestBalancerset(t *testing.T) {
	bs := newBalancerset(nil)
	tests := []struct {
		policy   BalancePolicy
		balancer balance.LoadBalancer
	}{
		{
			policy:   RandomBalancer,
			balancer: balance.NewRandomBalancer(nil),
		},
		{
			policy:   RoundRobinBalancer,
			balancer: balance.NewRoundRobinBalancer(nil),
		},
		{
			policy:   HashBalancer,
			balancer: balance.NewHashBalancer(nil, nil),
		},
		{
			policy:   NodeBalancer,
			balancer: balance.NewNodeBalancer(nil),
		},
	}
	for i, tt := range tests {
		if got, want := reflect.TypeOf(bs.getLoadBalancer(tt.policy)).Elem(), reflect.TypeOf(tt.balancer).Elem(); got != want {
			t.Errorf("%d: %v != %v", i, got, want)
		} else {
			t.Logf("%d: %v == %v", i, got, want)
		}
	}
}
