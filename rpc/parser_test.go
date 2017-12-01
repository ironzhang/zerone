package rpc

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestIsExported(t *testing.T) {
	tests := []struct {
		name   string
		expect bool
	}{
		{name: "", expect: false},
		{name: "A", expect: true},
		{name: "Aa", expect: true},
		{name: "Aaaaaa", expect: true},
		{name: "a", expect: false},
		{name: "aA", expect: false},
		{name: "aAAAAA", expect: false},
		{name: "你好", expect: false},
		{name: "A你好", expect: true},
		{name: "你好A", expect: false},
		{name: "_", expect: false},
		{name: "_A", expect: false},
		{name: "A_", expect: true},
	}
	for _, tt := range tests {
		if got, want := isExported(tt.name), tt.expect; got != want {
			t.Errorf("%q: %v != %v", tt.name, got, want)
		} else {
			t.Logf("%q: %v", tt.name, got)
		}
	}
}

func TestIsExportedOrBuiltinType(t *testing.T) {
	type a struct{}
	type A struct{}
	tests := []struct {
		typ    reflect.Type
		expect bool
	}{
		{typ: reflect.TypeOf(""), expect: true},
		{typ: reflect.TypeOf(1), expect: true},
		{typ: reflect.TypeOf(1.0), expect: true},
		{typ: reflect.TypeOf(a{}), expect: false},
		{typ: reflect.TypeOf(&a{}), expect: false},
		{typ: reflect.TypeOf(A{}), expect: true},
		{typ: reflect.TypeOf(&A{}), expect: true},
	}
	for i, tt := range tests {
		if got, want := isExportedOrBuiltinType(tt.typ), tt.expect; got != want {
			t.Errorf("case%d: %v: %v != %v", i, tt.typ.String(), got, want)
		} else {
			t.Logf("case%d: %v: %v", i, tt.typ.String(), got)
		}
	}
}

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

type args struct {
	A, B int
}

type reply struct {
	C int
}

type correct struct{}

func (correct) Test00(context.Context, int, *int) error {
	return nil
}

func (correct) Test01(context.Context, int, *string) error {
	return nil
}

func (correct) Test02(context.Context, int, *Reply) error {
	return nil
}

func (correct) Test03(context.Context, int, interface{}) error {
	return nil
}

func (correct) Test10(context.Context, *int, *int) error {
	return nil
}

func (correct) Test11(context.Context, Args, *int) error {
	return nil
}

func (correct) Test12(context.Context, *Args, *int) error {
	return nil
}

func (correct) Test13(context.Context, interface{}, *int) error {
	return nil
}

func (correct) Test20(Context, int, *int) error {
	return nil
}

type ill struct{}

func (ill) Test00() {
}

func (ill) Test01(context.Context, int, *int, int) (error, error) {
	return nil, nil
}

func (ill) Test10(int, int, *int) *error {
	return nil
}

func (ill) Test11(interface{}, int, *int) int {
	return 0
}

func (ill) Test20(context.Context, args, *int) bool {
	return false
}

func (ill) Test21(context.Context, *args, *int) {
}

func (ill) Test31(context.Context, int, int) {
}

func (ill) Test32(context.Context, int, *args) {
}

func TestCheckInsCorrect(t *testing.T) {
	var a correct
	typ := reflect.TypeOf(a)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		in0, in1, in2, in3, err := checkIns(m)
		if err != nil {
			t.Fatalf("checkIns: %v", err)
		}
		if got, want := in0, reflect.TypeOf(a); got != want {
			t.Fatalf("%v: in0: %v != %v", m.Name, got, want)
		}
		if got, want := in1, m.Type.In(1); got != want {
			t.Fatalf("%v: in1: %v != %v", m.Name, got, want)
		}
		if got, want := in2, m.Type.In(2); got != want {
			t.Fatalf("%v: in2: %v != %v", m.Name, got, want)
		}
		if got, want := in3, m.Type.In(3); got != want {
			t.Fatalf("%v: in3: %v != %v", m.Name, got, want)
		}
		t.Log(in0, in1, in2, in3)
	}
}

func TestCheckInsIll(t *testing.T) {
	var a ill
	typ := reflect.TypeOf(a)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if _, _, _, _, err := checkIns(m); err == nil {
			t.Fatalf("%v: checkIns return nil error", m.Name)
		} else {
			t.Logf("checkIns: %v", err)
		}
	}
}

func TestCheckOutsCorrect(t *testing.T) {
	var a correct
	typ := reflect.TypeOf(a)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if err := checkOuts(m); err != nil {
			t.Fatalf("checkOuts: %v", err)
		}
	}
}

func TestCheckOutsIll(t *testing.T) {
	var a ill
	typ := reflect.TypeOf(a)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if err := checkOuts(m); err == nil {
			t.Fatalf("%v: checkOuts return nil error", m.Name)
		} else {
			t.Logf("checkOuts: %v", err)
		}
	}
}
