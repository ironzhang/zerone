package balance

import (
	"strconv"
	"testing"

	"github.com/ironzhang/zerone/route"
)

type TestTB struct {
}

func (TestTB) ListEndpoints() []route.Endpoint {
	return []route.Endpoint{
		{"0", "tcp", "localhost:2000", 0.0},
		{"1", "tcp", "localhost:2001", 0.1},
		{"2", "tcp", "localhost:2002", 0.2},
	}
}

func RunLoadBalancerTests(t *testing.T, b route.LoadBalancer, name string, n int) {
	for i := 0; i < n; i++ {
		ep, err := b.GetEndpoint([]byte(strconv.Itoa(i)))
		if err != nil {
			t.Errorf("%s: GetEndpoint: %v", name, err)
		} else {
			t.Logf("%s: GetEndpoint: ep=%v", name, ep)
		}
	}
}

func TestRandomBalancer(t *testing.T) {
	b := NewRandomBalancer(TestTB{})
	RunLoadBalancerTests(t, b, "RandomBalancer", 10)
}

func TestRoundRobinBalancer(t *testing.T) {
	b := NewRoundRobinBalancer(TestTB{})
	RunLoadBalancerTests(t, b, "RoundRobinBalancer", 10)
}

func TestHashBalancer(t *testing.T) {
	b := NewHashBalancer(TestTB{}, nil)
	RunLoadBalancerTests(t, b, "HashBalancer", 10)
}

func TestNodeBalancer(t *testing.T) {
	b := NewNodeBalancer(TestTB{})
	RunLoadBalancerTests(t, b, "NodeBalancer", 3)
}
