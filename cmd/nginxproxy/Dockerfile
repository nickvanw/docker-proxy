FROM alpine:3.3

# Add an init system so we can run nginx and nginxproxy
ADD https://github.com/just-containers/s6-overlay/releases/download/v1.17.2.0/s6-overlay-amd64.tar.gz /tmp/
RUN gunzip -c /tmp/s6-overlay-amd64.tar.gz | tar -xf - -C /

# Add nginx
RUN apk add --update nginx

# Copy over nginxproxy
ADD nginxproxy /nginxproxy

# Add our service definitions
ADD init /etc/services.d/

# Add our nginx.conf
ADD nginx/nginx.conf /etc/nginx/nginx.conf

# Expose web ports
EXPOSE 80 443

# Enter into our init system
ENTRYPOINT /init