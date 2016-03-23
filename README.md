# docker-proxy [![GoDoc](https://godoc.org/github.com/nickvanw/docker-proxy?status.svg)](https://godoc.org/github.com/nickvanw/docker-proxy) [![Build Status](http://drone.dev.nvw.io/api/badges/nickvanw/docker-proxy/status.svg)](http://drone.dev.nvw.io/nickvanw/docker-proxy) [![Go Report Card](https://goreportcard.com/badge/github.com/nickvanw/docker-proxy)](https://goreportcard.com/report/github.com/nickvanw/docker-proxy) 

The Docker Proxy project consists of a few components. The end goal is to create a system for reverse proxying containers automatically as they are created and destroyed, allowing the end user to map a single endpoint to multiple containers. The project was created with Docker 1.10 in mind, meaning it supports overlay networks as a first class citizen. The first iteration currently does this using nginx, but a haproxy backend may also be in the works in the future.

The concept, as well as the environment variables to direct it, was taken from [jwilder/nginx-proxy](https://github.com/jwilder/nginx-proxy). Unfortunately, I found that I wanted a few things that it did not provide:

* More introspection into the current websites and containers that are mapped
* Production stability
* The ability to watch multiple Docker engines for change, which is very useful in a Swarm environment where the manager does not always report events. 

The project is currently a work in progress, but PRs are always welcome, as well as Issues/Feature Requests.


## Building

TBD

## Configuring

TBD

## Deploying

TBD