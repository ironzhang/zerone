package route

import (
	"reflect"
	"testing"
)

func TestTable(t *testing.T) {
	tb := NewTable()
	endpoints := []Endpoint{
		{"0", "localhost:2000", 0},
		{"1", "localhost:2001", 1},
		{"2", "localhost:2001", 2},
	}
	tb.AddEndpoints(endpoints...)

	if got, want := tb.ListEndpoints(), endpoints; !reflect.DeepEqual(got, want) {
		t.Errorf("endpoints: %v != %v", got, want)
	}

	tb.RemoveEndpoints("0")
	if got, want := tb.ListEndpoints(), endpoints[1:]; !reflect.DeepEqual(got, want) {
		t.Errorf("endpoints: %v != %v", got, want)
	}
}
