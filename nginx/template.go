package nginx

var nginxTemplate = `
# Pass along X-Forwarded-Proto if we have it, otherwise 
# pass along the scheme of the request
map $http_x_forwarded_proto $proxy_x_forwarded_proto {
  default $http_x_forwarded_proto;
  ''      $scheme;
}

# if we get an Upgrade, set the connection to "upgrade", or delete anything we've seen
map $http_upgrade $proxy_connection {
  default upgrade;
  '' close;
}

gzip_types text/plain text/css application/javascript application/json application/x-javascript text/xml application/xml application/xml+rss text/javascript;

log_format vhost '$host $remote_addr - $remote_user [$time_local] '
                 '"$request" $status $body_bytes_sent '
                 '"$http_referer" "$http_user_agent"';

access_log off;

# HTTP 1.1 support
proxy_http_version 1.1;
proxy_buffering off;
proxy_set_header Host $http_host;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection $proxy_connection;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $proxy_x_forwarded_proto;

# Set the default server
server {
	server_name _;
	listen 80;
	return 503;
}

{{ range . }}
{{ $ID := .ID }}
upstream {{ $ID }} {
	## Network: {{ .Contact.Network }}
	server {{ .Contact.Address }}:{{ .Contact.Port }};
}
{{ range .Hosts }}
server {
	server_name {{ . }};
	listen 80;
	location / {
		proxy_pass http://{{ $ID }};
	}
}
{{ end }}

{{ end }}
`
