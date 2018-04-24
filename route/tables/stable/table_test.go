package stable

import (
	"reflect"
	"testing"

	"github.com/ironzhang/zerone/route"
)

func TestTable(t *testing.T) {
	tests := []struct {
		ins  []route.Endpoint
		outs []route.Endpoint
	}{
		{
			ins: []route.Endpoint{
				{"0", "localhost:10000", 0.0},
				{"1", "localhost:10001", 0.11},
				{"2", "localhost:10002", 0.222},
			},
			outs: []route.Endpoint{
				{"0", "localhost:10000", 0.0},
				{"1", "localhost:10001", 0.11},
				{"2", "localhost:10002", 0.222},
			},
		},
		{
			ins: []route.Endpoint{
				{"1", "localhost:10001", 0.11},
				{"0", "localhost:10000", 0.0},
				{"2", "localhost:10002", 0.222},
			},
			outs: []route.Endpoint{
				{"0", "localhost:10000", 0.0},
				{"1", "localhost:10001", 0.11},
				{"2", "localhost:10002", 0.222},
			},
		},
	}
	for i, tt := range tests {
		tb := NewTable(tt.ins)
		if got, want := tb.ListEndpoints(), tt.outs; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: %v != %v", i, got, want)
		} else {
			t.Logf("%d: got=%v", i, got)
		}
	}
}
