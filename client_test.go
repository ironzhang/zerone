package zerone

import (
	"context"
	"testing"
	"time"

	"github.com/ironzhang/zerone/examples/rpc/arith"
	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/tables/stable"
)

func TestClient(t *testing.T) {
	tb := stable.NewTable([]route.Endpoint{
		{"0", "tcp", "localhost:10000", 0},
	})
	c := NewClient("test", tb).WithFailPolicy(NewFailtry(3, time.Second, 2*time.Second))
	defer c.Close()

	res := 0
	args := arith.Args{1, 2}
	err := c.Call(context.Background(), "Arith.Add", nil, args, &res)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res != args.A+args.B {
		t.Errorf("res(%d) != A(%d) + B(%d)", res, args.A, args.B)
	} else {
		t.Logf("res(%d) == A(%d) + B(%d)", res, args.A, args.B)
	}
}
