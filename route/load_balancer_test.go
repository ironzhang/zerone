package route

import (
	"strconv"
	"testing"
)

type TestTable struct {
}

func (t TestTable) ListEndpoints() []Endpoint {
	return []Endpoint{
		{"localhost:2000", 0.1},
		{"localhost:2001", 0.1},
		{"localhost:2002", 0.1},
	}
}

func RunLoadBalancerTests(t *testing.T, b LoadBalancer, name string, n int) {
	for i := 0; i < 10; i++ {
		ep, err := b.GetEndpoint([]byte(strconv.Itoa(i)))
		if err != nil {
			t.Errorf("%s: GetEndpoint: %v", name, err)
		} else {
			t.Logf("%s: GetEndpoint: ep=%v", name, ep)
		}
	}
}

func TestRandomBalancer(t *testing.T) {
	b := NewRandomBalancer(TestTable{})
	RunLoadBalancerTests(t, b, "RandomBalancer", 10)
}

func TestRoundRobinBalancer(t *testing.T) {
	b := NewRoundRobinBalancer(TestTable{})
	RunLoadBalancerTests(t, b, "RoundRobinBalancer", 10)
}

func TestHashBalancer(t *testing.T) {
	b := NewHashBalancer(TestTable{}, nil)
	RunLoadBalancerTests(t, b, "HashBalancer", 10)
}
