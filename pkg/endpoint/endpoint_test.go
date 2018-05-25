package endpoint

import "testing"

func TestEndpointEqual(t *testing.T) {
	tests := []struct {
		a    Endpoint
		b    interface{}
		want bool
	}{
		{
			a:    Endpoint{Name: "n1", Net: "tcp", Addr: "localhost:2000", Load: 0},
			b:    &Endpoint{Name: "n1", Net: "tcp", Addr: "localhost:2000", Load: 0},
			want: true,
		},
		{
			a:    Endpoint{Name: "n1", Net: "tcp", Addr: "localhost:2000", Load: 0.0},
			b:    &Endpoint{Name: "n1", Net: "tcp", Addr: "localhost:2000", Load: 0.1},
			want: false,
		},
	}
	for i, tt := range tests {
		if got, want := tt.a.Equal(tt.b), tt.want; got != want {
			t.Errorf("%d: got %v, want %v", i, got, want)
		} else {
			t.Logf("%d: got %v", i, got)
		}
	}
}
