package nginx

var nginxHeader = `
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

# Set appropriate X-Forwarded-Ssl header
map $scheme $proxy_x_forwarded_ssl {
  default off;
  https on;
}

gzip_types text/plain text/css application/javascript application/json application/x-javascript text/xml application/xml application/xml+rss text/javascript;


{{ if .Syslog }}
log_format loggly '$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" - $request_time X-Forwarded-For=$http_x_forwarded_for Host=$host';
error_log syslog:server={{.Syslog}};
access_log syslog:server={{.Syslog}} loggly;
{{else}}
access_log off;
{{end}}

# HTTP 1.1 support
proxy_http_version 1.1;
proxy_buffering off;
proxy_set_header Host $http_host;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection $proxy_connection;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $proxy_x_forwarded_proto;
proxy_set_header X-Forwarded-Ssl $proxy_x_forwarded_ssl;

# Set the default server
server {
	server_name _;
	listen 80;
	return 503;
}
`

var nginxUpstream = `
upstream {{ .ID }} {
	## Network: {{ .Contact.Network }}
	server {{ .Contact.Address }}:{{ .Contact.Port }};
}
`

var nginxWithSSL = `
server {
	server_name {{ .Host }};
	listen 80;
	return 301 https://$host$request_uri;
}

server {
	server_name {{ .Host }};
	listen 443 ssl;

	ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	ssl_ciphers "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:ECDHE-RSA-DES-CBC3-SHA:ECDHE-ECDSA-DES-CBC3-SHA:EDH-RSA-DES-CBC3-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA";
	ssl_prefer_server_ciphers on;
	ssl_session_timeout 5m;
	ssl_session_cache shared:SSL:50m;

	ssl_certificate {{ .SSLPrefix }}.crt;
	ssl_certificate_key {{ .SSLPrefix }}.key;
	add_header Strict-Transport-Security "max-age=31536000";
	{{ template "config" .Config }}
	location / {
		proxy_pass http://{{ .ID }};
	}
}
`

var nginxNoSSL = `
server {
	server_name {{ .Host }};
	listen 80;
	{{ template "config" .Config }}
	location / {
		proxy_pass http://{{ .ID }};
	}
}
`

var nginxOptions = `{{ define "config" }}{{ range $key, $value := . }}{{ $key }} {{ $value }};{{ end }} {{ end }}`
