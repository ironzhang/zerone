package zerone

import (
	"reflect"
	"testing"
)

func TestNewValue(t *testing.T) {
	type s struct {
		i8  int8
		i16 int16
		f32 float32
		f64 float64
	}
	values := []interface{}{
		nil,
		int(1),
		int8(2),
		float64(3.3),
		s{},
		&s{},
	}
	for i, value := range values {
		if got, want := reflect.TypeOf(newValue(value)), reflect.TypeOf(value); got != want {
			t.Errorf("%d: %v != %v", i, got, want)
		} else {
			t.Logf("%d: %v == %v", i, got, want)
		}
	}
}
