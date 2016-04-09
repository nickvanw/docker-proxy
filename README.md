# docker-proxy [![GoDoc](https://godoc.org/github.com/nickvanw/docker-proxy?status.svg)](https://godoc.org/github.com/nickvanw/docker-proxy) [![Build Status](http://drone.dev.nvw.io/api/badges/nickvanw/docker-proxy/status.svg)](http://drone.dev.nvw.io/nickvanw/docker-proxy) [![Go Report Card](https://goreportcard.com/badge/github.com/nickvanw/docker-proxy)](https://goreportcard.com/report/github.com/nickvanw/docker-proxy) 

The Docker Proxy project consists of a few components. The end goal is to create a system for reverse proxying containers automatically as they are created and destroyed, allowing the end user to map a single endpoint to multiple containers. The project was created with Docker 1.10+ in mind, meaning it supports overlay networks as a first class citizen. The first iteration currently does this using nginx, but a haproxy backend may also be in the works in the future.

The concept, as well as the environment variables used in some of the configuration, was taken from [jwilder/nginx-proxy](https://github.com/jwilder/nginx-proxy). Unfortunately, I found that I wanted a few things that it did not provide:

* More introspection into the current websites and containers that are mapped
* Production stability
* The ability to watch multiple Docker engines for change, which is very useful in a Swarm environment where the manager does not always report events. 

The project is currently a work in progress, but PRs are always welcome, as well as Issues/Feature Requests.


## Building

The project can be built and fetched with a `go get github.com/nickvanw/docker-proxy` and `go build` from `cmd/nginxproxy`

There is also a Dockerfile located in `cmd/nginxproxy` that will build an alpine container using the built binary in `cmd/nginxproxy`. There are some caveats to this:

* The Dockerfile is built using alpine, meaning there is no glibc - I recommend statically compiling with musl
* The Dockerfile uses the latest version of nginx in the alpine stable repository, which will sometimes lag behind the latest build

## Configuring

Configuration of the nginx proxy is done via flags and/or environment variables. The latest list of them, as well as a description of what they do, can be found by building the binary and running `nginxproxy -h`. A few of the more important variables:

* `DOCKER_HOST` should point to the unix or TCP endpoint where the proxy can pull the list of containers, as well as events from.
* `NGINX_CONF` is the file that the nginx proxy will write to
* `NGINX_RELOAD` is the command the nginx proxy will run to reload nginx when the containers are updated

The rest of the flags are useful for enabling additional functions, such as shipping nginx and proxy logs to syslog, or enabling sentry reporting for errors.

## Deploying

The strategies used to deploy this will change depending on how your Docker setup works. Currently, I'm running it with a 5 node swarm cluster with three of the nodes running this service, using round-robin DNS to split the load across them.

Each host has the Docker container deployed out with ports 80 and 443 bonded to their external addresses. They are all listening to events on the swarm master, as well as all of the swarm nodes. This causes a small amount of additional load, but the debounce usually results in the container list only being pulled once, and there is a significantly lower change of events getting dropped. 

## Usage

Using the service is designed to be as simple as possible - there is only one environment variable necessary on new containers for the proxy to begin working, `VIRTUAL_HOST`. If the container exposes more than one port, `VIRTUAL_PORT` is required as well, or the service will be unaware of where to proxy to, and will ignore the container. 

Example:

```
docker run -e VIRTUAL_HOST=something.nvw.io ...
```

If the DNS of `something.nvw.io` is pointed to the container(s) running the proxy, and the image exposes a single port, traffic will be forwarded seamlessly. 

### HTTPS

HTTPS is fully supported - a container with VIRTUAL_HOST=foo.bar.com should have a foo.bar.com.crt and foo.bar.com.key file in the certs directory (configured in the service). It will automatically redirect non-https access to the https location, and will set an HSTS header. 

### Arbitrary configuration options

Currently, there are a few environment variables that will map to nginx configuration directives. More are very easy to add, I just haven't had a use:

* `NGINX_CLIENT_MAX_BODY_SIZE` will set the max body size for the virtual host
* `AUTH_USER` and `AUTH_PASS` (with a configured htpasswd directory) will create an htpasswd file and use that for basic auth. AUTH_PASS will be printed verbatim into the file, so use the encryption scheme of choice, but I recommend bcrypt. 
