package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nickvanw/docker-proxy"
	"golang.org/x/net/context"
)

func main() {
	cfg := dockerproxy.DockerConfig{
		Addr: "tcp://192.168.99.100:2376",
		TLS:  true,
		Cert: "/Users/nick/.docker/machine/machines/default/cert.pem",
		CA:   "/Users/nick/.docker/machine/machines/default/ca.pem",
		Key:  "/Users/nick/.docker/machine/machines/default/key.pem",
	}
	m, err := dockerproxy.New(cfg)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx, time.Minute)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancel()
}
