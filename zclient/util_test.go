package zclient

import (
	"reflect"
	"testing"
)

func TestNewValuePtr(t *testing.T) {
	type s struct {
		i8  int8
		i16 int16
		f32 float32
		f64 float64
	}
	tests := []struct {
		ovalue interface{}
		nvalue interface{}
	}{
		{ovalue: nil, nvalue: nil},
		{ovalue: int(1), nvalue: new(int)},
		{ovalue: int8(2), nvalue: new(int8)},
		{ovalue: float64(3.3), nvalue: new(float64)},
		{ovalue: s{}, nvalue: new(s)},
		{ovalue: &s{}, nvalue: new(s)},
	}
	for i, tt := range tests {
		if got, want := reflect.TypeOf(newValuePtr(tt.ovalue)), reflect.TypeOf(tt.nvalue); got != want {
			t.Errorf("%d: %v != %v", i, got, want)
		} else {
			t.Logf("%d: %v == %v", i, got, want)
		}
	}
}
