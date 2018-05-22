package stable

import (
	"reflect"
	"testing"

	"github.com/ironzhang/pearls/config"
	"github.com/ironzhang/zerone/route"
)

func TestTable(t *testing.T) {
	tests := []struct {
		ins  []route.Endpoint
		outs []route.Endpoint
	}{
		{
			ins: []route.Endpoint{
				{"0", "tcp", "localhost:10000", 0.0},
				{"1", "tcp", "localhost:10001", 0.11},
				{"2", "tcp", "localhost:10002", 0.222},
			},
			outs: []route.Endpoint{
				{"0", "tcp", "localhost:10000", 0.0},
				{"1", "tcp", "localhost:10001", 0.11},
				{"2", "tcp", "localhost:10002", 0.222},
			},
		},
		{
			ins: []route.Endpoint{
				{"1", "udp", "localhost:10001", 0.11},
				{"0", "udp", "localhost:10000", 0.0},
				{"2", "udp", "localhost:10002", 0.222},
			},
			outs: []route.Endpoint{
				{"0", "udp", "localhost:10000", 0.0},
				{"1", "udp", "localhost:10001", 0.11},
				{"2", "udp", "localhost:10002", 0.222},
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

func TestLoadTables(t *testing.T) {
	config.Default = config.TOML

	wtables := Tables{
		"account": []route.Endpoint{
			{"0", "tcp", "localhost:10000", 0.0},
			{"1", "tcp", "localhost:10001", 0.11},
			{"2", "tcp", "localhost:10002", 0.222},
		},
		"logger": []route.Endpoint{
			{"0", "udp", "localhost:10000", 0.0},
			{"1", "udp", "localhost:10001", 0.11},
			{"2", "udp", "localhost:10002", 0.222},
		},
	}
	if err := config.WriteToFile("example.conf", wtables); err != nil {
		t.Fatalf("write to file: %v", err)
	}

	rtables, err := LoadTables("example.conf")
	if err != nil {
		t.Fatalf("load table: %v", err)
	}
	if got, want := rtables, wtables; !reflect.DeepEqual(got, want) {
		t.Fatalf("tables: got %v, want %v", got, want)
	} else {
		t.Logf("tables: got %v", got)
	}
}

func TestTablesLookup(t *testing.T) {
	tables := Tables{
		"account": []route.Endpoint{
			{"0", "tcp", "localhost:10000", 0.0},
			{"1", "tcp", "localhost:10001", 0.11},
			{"2", "tcp", "localhost:10002", 0.222},
		},
		"logger": []route.Endpoint{
			{"0", "udp", "localhost:10000", 0.0},
			{"1", "udp", "localhost:10001", 0.11},
			{"2", "udp", "localhost:10002", 0.222},
		},
	}

	for svc, eps := range tables {
		tb, err := tables.Lookup(svc)
		if err != nil {
			t.Fatalf("%s: Lookup: %v", svc, err)
		}
		if got, want := tb.ListEndpoints(), eps; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: endpoints: got %v, want %v", svc, got, want)
		} else {
			t.Logf("%s: endpoints: got %v", svc, got)
		}
	}
}
