package zclient

import (
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/ironzhang/zerone/pkg/balance"
	"github.com/ironzhang/zerone/pkg/endpoint"
	"github.com/ironzhang/zerone/pkg/route/stable"
	"github.com/ironzhang/zerone/rpc"
)

func TestFailtry(t *testing.T) {
	tb := stable.NewTable([]endpoint.Endpoint{
		{"0", "tcp", "localhost:10000", 0},
		{"1", "tcp", "localhost:10001", 0},
	})

	var (
		sleep time.Duration
		docnt int
		addrs []string
	)
	timeSleep = func(d time.Duration) { sleep += d }

	tests := []struct {
		try   int
		min   time.Duration
		max   time.Duration
		err   error
		docnt int
		sleep time.Duration
		addrs []string
	}{
		{
			try:   0,
			min:   0,
			max:   0,
			err:   nil,
			sleep: 0,
			docnt: 1,
			addrs: []string{
				"tcp://localhost:10000",
			},
		},
		{
			try:   0,
			min:   0,
			max:   0,
			err:   io.EOF,
			sleep: 0,
			docnt: 1,
			addrs: []string{
				"tcp://localhost:10000",
			},
		},
		{
			try:   2,
			min:   time.Second,
			max:   0,
			err:   io.EOF,
			sleep: time.Second,
			docnt: 2,
			addrs: []string{
				"tcp://localhost:10000",
				"tcp://localhost:10000",
			},
		},
		{
			try:   3,
			min:   time.Second,
			max:   0,
			err:   io.EOF,
			sleep: 2 * time.Second,
			docnt: 3,
			addrs: []string{
				"tcp://localhost:10000",
				"tcp://localhost:10000",
				"tcp://localhost:10000",
			},
		},
		{
			try:   3,
			min:   time.Second,
			max:   2 * time.Second,
			err:   io.EOF,
			sleep: 3 * time.Second,
			docnt: 3,
			addrs: []string{
				"tcp://localhost:10000",
				"tcp://localhost:10000",
				"tcp://localhost:10000",
			},
		},
		{
			try:   3,
			min:   2 * time.Second,
			max:   0,
			err:   io.EOF,
			sleep: 2 * time.Second,
			docnt: 3,
			addrs: []string{
				"tcp://localhost:10000",
				"tcp://localhost:10000",
				"tcp://localhost:10000",
			},
		},
		{
			try:   3,
			min:   2 * time.Second,
			max:   4 * time.Second,
			err:   io.EOF,
			sleep: 6 * time.Second,
			docnt: 3,
			addrs: []string{
				"tcp://localhost:10000",
				"tcp://localhost:10000",
				"tcp://localhost:10000",
			},
		},
		{
			try:   3,
			min:   4 * time.Second,
			max:   2 * time.Second,
			err:   io.EOF,
			sleep: 4 * time.Second,
			docnt: 3,
			addrs: []string{
				"tcp://localhost:10000",
				"tcp://localhost:10000",
				"tcp://localhost:10000",
			},
		},
	}
	for i, tt := range tests {
		sleep = 0
		docnt = 0
		addrs = nil
		do := func(net, addr string) (*rpc.Call, error) {
			docnt++
			addrs = append(addrs, fmt.Sprintf("%s://%s", net, addr))
			return nil, tt.err
		}

		lb := balance.NewRoundRobinBalancer(tb)
		f := NewFailtry(tt.try, tt.min, tt.max)
		f.execute(lb, nil, do)

		if got, want := docnt, tt.docnt; got != want {
			t.Errorf("%d: docnt: %v != %v", i, got, want)
		}
		if got, want := sleep, tt.sleep; got != want {
			t.Errorf("%d: sleep: %v != %v", i, got, want)
		}
		if got, want := addrs, tt.addrs; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: addrs: %v != %v", i, got, want)
		}
	}
}

func TestFailover(t *testing.T) {
	tb := stable.NewTable([]endpoint.Endpoint{
		{"0", "tcp", "localhost:10000", 0},
		{"1", "tcp", "localhost:10001", 0},
	})

	var (
		docnt int
		addrs []string
	)

	tests := []struct {
		try   int
		err   error
		docnt int
		addrs []string
	}{
		{
			try:   0,
			err:   nil,
			docnt: 1,
			addrs: []string{
				"tcp://localhost:10000",
			},
		},
		{
			try:   2,
			err:   nil,
			docnt: 1,
			addrs: []string{
				"tcp://localhost:10000",
			},
		},
		{
			try:   0,
			err:   io.EOF,
			docnt: 1,
			addrs: []string{
				"tcp://localhost:10000",
			},
		},
		{
			try:   2,
			err:   io.EOF,
			docnt: 2,
			addrs: []string{
				"tcp://localhost:10000",
				"tcp://localhost:10001",
			},
		},
		{
			try:   3,
			err:   io.EOF,
			docnt: 3,
			addrs: []string{
				"tcp://localhost:10000",
				"tcp://localhost:10001",
				"tcp://localhost:10000",
			},
		},
	}
	for i, tt := range tests {
		docnt = 0
		addrs = nil
		do := func(net, addr string) (*rpc.Call, error) {
			docnt++
			addrs = append(addrs, fmt.Sprintf("%s://%s", net, addr))
			return nil, tt.err
		}

		lb := balance.NewRoundRobinBalancer(tb)
		f := NewFailover(tt.try)
		f.execute(lb, nil, do)

		if got, want := docnt, tt.docnt; got != want {
			t.Errorf("%d: docnt: %v != %v", i, got, want)
		}
		if got, want := addrs, tt.addrs; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: addrs: %v != %v", i, got, want)
		}
	}
}
