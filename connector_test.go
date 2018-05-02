package zerone

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestConnectorDial(t *testing.T) {
	c := newConnector("", nil, 0)

	type point struct {
		key, net, addr string
	}

	tests := []struct {
		p1   point
		p2   point
		same bool
	}{
		{
			p1: point{
				key:  "1",
				net:  "tcp",
				addr: "localhost:10000",
			},
			p2: point{
				key:  "1",
				net:  "tcp",
				addr: "localhost:10000",
			},
			same: true,
		},
		{
			p1: point{
				key:  "1",
				net:  "tcp",
				addr: "localhost:10000",
			},
			p2: point{
				key:  "2",
				net:  "tcp",
				addr: "localhost:10000",
			},
			same: false,
		},
	}
	for i, tt := range tests {
		c1, err := c.dial(tt.p1.key, tt.p1.net, tt.p1.addr)
		if err != nil {
			t.Fatalf("%d: dial: %v", i, err)
		}
		c2, err := c.dial(tt.p2.key, tt.p2.net, tt.p2.addr)
		if err != nil {
			t.Fatalf("%d: dial: %v", i, err)
		}
		if got, want := c1 == c2, tt.same; got != want {
			t.Errorf("%d: got %v want %v, c1=%p, c2=%p", i, got, want, c1, c2)
		}
	}
}

func BenchmarkConnectorDial(b *testing.B) {
	b.SetParallelism(50)
	c := newConnector("", nil, 0)
	b.RunParallel(func(pb *testing.PB) {
		key := fmt.Sprint(rand.Int())
		for pb.Next() {
			_, err := c.dial(key, "tcp", "localhost:10000")
			if err != nil {
				b.Fatalf("dial: %v", err)
			}
		}
	})
	//fmt.Printf("client's num: %d\n", len(c.clients))
}
