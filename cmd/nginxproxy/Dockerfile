FROM alpine:3.4

# Add an init system so we can run nginx and nginxproxy
ADD https://github.com/just-containers/s6-overlay/releases/download/v1.17.2.0/s6-overlay-amd64.tar.gz /tmp/
RUN gunzip -c /tmp/s6-overlay-amd64.tar.gz | tar -xf - -C /

# Add nginx
RUN apk add --update nginx

# Make htpasswd directory
RUN mkdir -p /etc/nginx/htpasswd

# Copy over nginxproxy
ADD nginxproxy /nginxproxy

# Add our service definitions
ADD init /etc/services.d/

# Add our nginx.conf
ADD nginx/nginx.conf /etc/nginx/nginx.conf

# Expose web ports
EXPOSE 80 443

# Environment variables guaranteed with this container
ENV NGINX_RELOAD_CMD "nginx -s reload"
ENV NGINX_HTPASSWD_DIR "/etc/nginx/htpasswd"
ENV NGINX_CONF "/etc/nginx/conf.d/nginxproxy.conf"

# Enter into our init system
ENTRYPOINT /init
