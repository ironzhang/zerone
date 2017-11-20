package rpc

import (
	"context"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var (
	typeOfError        = reflect.TypeOf((*error)(nil)).Elem()
	typeOfContext      = reflect.TypeOf((*context.Context)(nil)).Elem()
	typeOfNilInterface = reflect.TypeOf((*interface{})(nil)).Elem()
)

// Is this an exported - upper case - name?
func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// Method needs four ins: receiver, context.Context, *args, *reply.
func checkIns(m reflect.Method) (in0, in1, in2, in3 reflect.Type, err error) {
	mtype := m.Type
	if mtype.NumIn() != 4 {
		err = fmt.Errorf("method %s has wrong number of ins: %d", m.Name, mtype.NumIn())
		return
	}
	in0 = mtype.In(0)
	in1 = mtype.In(1)
	if !in1.Implements(typeOfContext) {
		err = fmt.Errorf("method %s context type not implements context.Context: %s", m.Name, in1)
		return
	}
	in2 = mtype.In(2)
	if !isExportedOrBuiltinType(in2) {
		err = fmt.Errorf("method %s args type not exported: %s", m.Name, in2)
		return
	}
	in3 = mtype.In(3)
	if in3.Kind() != reflect.Ptr && in3 != typeOfNilInterface {
		err = fmt.Errorf("method %s reply type not a pointer or interface{}: %s", m.Name, in3)
		return
	}
	if !isExportedOrBuiltinType(in3) {
		err = fmt.Errorf("method %s reply type not exported: %s", m.Name, in3)
		return
	}
	return
}

// The return type of the method must be error.
func checkOuts(m reflect.Method) error {
	mtype := m.Type
	if mtype.NumOut() != 1 {
		return fmt.Errorf("method %s has wrong number of outs: %d", m.Name, mtype.NumOut())
	}
	if out0 := mtype.Out(0); out0 != typeOfError {
		return fmt.Errorf("method %s returns %s not error", m.Name, out0)
	}
	return nil
}

type method struct {
	meth  reflect.Method
	args  reflect.Type
	reply reflect.Type
}

func parseMethod(m reflect.Method) (*method, error) {
	_, _, args, reply, err := checkIns(m)
	if err != nil {
		return nil, err
	}
	if err = checkOuts(m); err != nil {
		return nil, err
	}
	return &method{meth: m, args: args, reply: reply}, nil
}

type service struct {
	rcvr    reflect.Value
	methods map[string]*method
}

func parseService(rcvr reflect.Value) (*service, error) {
	typ := rcvr.Type()
	methods := make(map[string]*method)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if m.PkgPath != "" {
			continue
		}
		mt, err := parseMethod(m)
		if err != nil {
			return nil, err
		}
		methods[m.Name] = mt
	}
	return &service{rcvr: rcvr, methods: methods}, nil
}