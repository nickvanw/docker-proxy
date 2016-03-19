package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/nickvanw/docker-proxy"
	"github.com/nickvanw/docker-proxy/nginx"
	"golang.org/x/net/context"
)

func main() {
	app := cli.NewApp()
	app.Name = "nginxproxy"
	app.Action = realMain
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "docker.leader, l",
			Value:  "",
			Usage:  "Leading Docker server to pull containers from",
			EnvVar: "DOCKER_HOST",
		},
		cli.BoolFlag{
			Name:   "docker.tls",
			Usage:  "Connect to Docker with TLS",
			EnvVar: "DOCKER_TLS_VERIFY",
		},
		cli.StringFlag{
			Name:   "docker.certs",
			Value:  "",
			Usage:  "Location of Docker TLS certs",
			EnvVar: "DOCKER_CERT_PATH",
		},
		cli.StringFlag{
			Name:   "followers, f",
			Value:  "",
			Usage:  "Follower Docker servers to poll for events",
			EnvVar: "DOCKER_FOLLOWERS",
		},
		cli.StringFlag{
			Name:   "nginx.conf",
			Value:  "/etc/nginx/conf.d/nginxproxy.conf",
			Usage:  "nginx config to write",
			EnvVar: "NGINX_CONF",
		},
		cli.StringFlag{
			Name:   "nginx.certs",
			Value:  "/opt/nginxssl",
			Usage:  "Location of nginx SSL certs",
			EnvVar: "NGINX_CERT_PATH",
		},
		cli.StringFlag{
			Name:   "nginx.reload",
			Value:  "",
			Usage:  "command to reload nginx",
			EnvVar: "NGINX_RELOAD_CMD",
		},
	}
	app.Run(os.Args)
}

func realMain(c *cli.Context) {
	cfg := dockerproxy.DockerConfig{
		Leader: c.String("docker.leader"),
		TLS:    c.Bool("docker.tls"),
	}

	if followers := c.String("followers"); followers != "" {
		cfg.Watchers = strings.Split(followers, ",")
	}

	if dir := c.String("docker.certs"); dir != "" {
		cfg.Cert = filepath.Join(dir, "cert.pem")
		cfg.CA = filepath.Join(dir, "ca.pem")
		cfg.Key = filepath.Join(dir, "key.pem")
	}

	m, err := dockerproxy.New(cfg)
	if err != nil {
		log.Fatalf("error creating new docker proxy: %s", err)
	}

	nginx, err := nginx.New(c.String("nginx.certs"),
		c.String("nginx.conf"), c.String("nginx.reload"))
	if err != nil {
		log.Fatalf("error creating nginx watcher: %s", err)
	}
	m.Register(nginx)

	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx, 10*time.Minute)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancel()
}
