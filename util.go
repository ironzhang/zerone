package zerone

import "reflect"

func newValue(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	t := reflect.TypeOf(a)
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface()
	} else {
		return reflect.New(t).Elem().Interface()
	}
}
