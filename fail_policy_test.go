package zerone

import (
	"io"
	"testing"
	"time"

	"github.com/ironzhang/zerone/route"
	"github.com/ironzhang/zerone/route/balance"
	"github.com/ironzhang/zerone/route/tables/stable"
	"github.com/ironzhang/zerone/rpc"
)

func TestFailtry(t *testing.T) {
	tb := stable.NewTable([]route.Endpoint{
		{"0", "tcp", "localhost:10000", 0},
	})
	lb := balance.NewRandomBalancer(tb)

	var (
		sleep time.Duration
		docnt int
	)
	timeSleep = func(d time.Duration) { sleep += d }

	tests := []struct {
		try   int
		min   time.Duration
		max   time.Duration
		err   error
		docnt int
		sleep time.Duration
	}{
		{
			try:   0,
			min:   0,
			max:   0,
			err:   nil,
			sleep: 0,
			docnt: 1,
		},
		{
			try:   0,
			min:   0,
			max:   0,
			err:   io.EOF,
			sleep: 0,
			docnt: 1,
		},
		{
			try:   2,
			min:   time.Second,
			max:   0,
			err:   io.EOF,
			sleep: time.Second,
			docnt: 2,
		},
		{
			try:   3,
			min:   time.Second,
			max:   0,
			err:   io.EOF,
			sleep: 2 * time.Second,
			docnt: 3,
		},
		{
			try:   3,
			min:   time.Second,
			max:   2 * time.Second,
			err:   io.EOF,
			sleep: 3 * time.Second,
			docnt: 3,
		},
		{
			try:   3,
			min:   2 * time.Second,
			max:   0,
			err:   io.EOF,
			sleep: 2 * time.Second,
			docnt: 3,
		},
		{
			try:   3,
			min:   2 * time.Second,
			max:   4 * time.Second,
			err:   io.EOF,
			sleep: 6 * time.Second,
			docnt: 3,
		},
		{
			try:   3,
			min:   4 * time.Second,
			max:   2 * time.Second,
			err:   io.EOF,
			sleep: 4 * time.Second,
			docnt: 3,
		},
	}
	for i, tt := range tests {
		sleep = 0
		docnt = 0
		do := func(net, addr string) (*rpc.Call, error) {
			docnt++
			return nil, tt.err
		}

		f := NewFailtry(tt.try, tt.min, tt.max)
		f.execute(lb, nil, do)

		if got, want := docnt, tt.docnt; got != want {
			t.Errorf("%d: docnt: %v != %v", i, got, want)
		}
		if got, want := sleep, tt.sleep; got != want {
			t.Errorf("%d: sleep: %v != %v", i, got, want)
		}
	}
}
