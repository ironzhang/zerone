package rpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/ironzhang/x-pearls/log"
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
	method reflect.Method
	args   reflect.Type
	reply  reflect.Type
}

func parseMethod(m reflect.Method) (*method, error) {
	_, _, args, reply, err := checkIns(m)
	if err != nil {
		return nil, err
	}
	if err = checkOuts(m); err != nil {
		return nil, err
	}
	return &method{method: m, args: args, reply: reply}, nil
}

func suitableMethods(typ reflect.Type, reportErr bool) map[string]*method {
	methods := make(map[string]*method)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if m.PkgPath != "" {
			continue
		}
		meth, err := parseMethod(m)
		if err != nil {
			if reportErr {
				log.Debugf("parse method: %v", err)
			}
			continue
		}
		methods[m.Name] = meth
	}
	return methods
}

type class struct {
	name    string
	rcvr    reflect.Value
	methods map[string]*method
}

func parseClass(name string, rcvr reflect.Value) (*class, error) {
	typ := rcvr.Type()
	methods := suitableMethods(typ, true)
	if len(methods) <= 0 {
		var str string
		methods = suitableMethods(reflect.PtrTo(typ), false)
		if len(methods) <= 0 {
			str = fmt.Sprintf("type %s has no exported methods of suitable type", typ.Name())
		} else {
			str = fmt.Sprintf("type %s has no exported methods of suitable type (hint: pass a pointer to value of that type)", typ.Name())
		}
		return nil, errors.New(str)
	}
	return &class{name: name, rcvr: rcvr, methods: methods}, nil
}

func isNilInterface(t reflect.Type) bool {
	return t == typeOfNilInterface
}
