package dockerproxy

import (
	"errors"
	"net/url"

	"github.com/fsouza/go-dockerclient"
)

var (
	ErrInvalidURL = errors.New("unknown URL scheme for docker endpoint")
)

func newDockerClient(addr string, tls bool, cert, ca, key string) (*docker.Client, error) {
	url, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	switch url.Scheme {
	case "unix":
		return docker.NewClient(addr)
	case "tcp":
		if tls {
			return docker.NewTLSClient(addr, cert, key, ca)
		}
		return docker.NewClient(addr)
	default:
		return nil, ErrInvalidURL
	}
}
