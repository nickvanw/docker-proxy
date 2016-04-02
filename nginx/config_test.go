package nginx

import (
	"reflect"
	"testing"
)

func TestEnvToDirective(t *testing.T) {
	tt := []struct {
		env     map[string]string
		wantEnv map[string]string
	}{
		{
			env:     map[string]string{"NGINX_CLIENT_MAX_BODY_SIZE": "0m"},
			wantEnv: map[string]string{"client_max_body_size": "0m"},
		},
		{
			env:     map[string]string{},
			wantEnv: map[string]string{},
		},
	}

	for _, v := range tt {
		got := envToDirectives(v.env)
		if !reflect.DeepEqual(v.wantEnv, got) {
			t.Fatalf("wanted: %#v, got: %#v", v.wantEnv, got)
		}
	}
}
