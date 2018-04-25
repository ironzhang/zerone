package zerone

import (
	"context"
	"testing"

	"github.com/ironzhang/zerone/examples/rpc/arith"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/tables/stable"
)

func TestClient(t *testing.T) {
	tb := stable.NewTable([]route.Endpoint{{"1", "localhost:10000", 0}})
	c := NewClient("test", tb)
	//c.SetTraceVerbose(1)

	res := 0
	args := arith.Args{1, 2}
	err := c.WithLoadBalancer(RoundRobinBalancer).Call(context.Background(), "Arith.Add", nil, args, &res)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res != args.A+args.B {
		t.Errorf("res(%d) != A(%d) + B(%d)", res, args.A, args.B)
	} else {
		t.Logf("res(%d) == A(%d) + B(%d)", res, args.A, args.B)
	}
}
