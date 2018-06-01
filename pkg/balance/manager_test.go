package balance

import "testing"

func TestManager(t *testing.T) {
	m := NewManager(TestTB{}, nil)

	names := []string{RandomBalancerName, RoundRobinBalancerName, HashBalancerName, NodeBalancerName}
	for i, name := range names {
		lb := m.GetLoadBalancer(name)
		if got, want := lb.Name(), name; got != want {
			t.Fatalf("%d: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: got %v", i, got)
		}
	}
	lb := m.GetLoadBalancer("default")
	if got, want := lb.Name(), RandomBalancerName; got != want {
		t.Fatalf("default: got %v, want %v", got, want)
	} else {
		t.Logf("default: got %v", got)
	}
}
