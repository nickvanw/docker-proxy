package dockerproxy

import (
	"reflect"
	"testing"
)

func TestKV(t *testing.T) {
	tt := []struct {
		strs    []string
		want    map[string]string
		wantErr error
	}{
		{
			strs:    []string{"a=b", "b=c", "d=d=d=d=d", "HELLO=WORLD"},
			want:    map[string]string{"a": "b", "b": "c", "d": "d=d=d=d", "HELLO": "WORLD"},
			wantErr: nil,
		},
		{
			strs:    []string{"a"},
			want:    nil,
			wantErr: ErrInvalidLine,
		},
	}

	for _, v := range tt {
		got, err := parseKV(v.strs)
		if err != v.wantErr {
			t.Fatalf("want err: %s, got: %s", v.wantErr, err)
		}
		if !reflect.DeepEqual(v.want, got) {
			t.Fatalf("want: %#v, got: %#v", v.want, got)
		}
	}
}
