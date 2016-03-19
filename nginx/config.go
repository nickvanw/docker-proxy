package nginx

var m = map[string]string{
	"NGINX_CLIENT_MAX_BODY_SIZE": "client_max_body_size",
}

func envToDirectives(env map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range m {
		if val, ok := env[k]; ok {
			out[v] = val
		}
	}
	return out
}
