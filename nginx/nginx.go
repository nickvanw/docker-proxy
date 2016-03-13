package nginx

import (
	"fmt"

	"github.com/nickvanw/docker-proxy"
)

type Server struct {
	SSLDir   string
	Reloader Reloader
}

func (s *Server) Start(sites []dockerproxy.Site) error {
	fmt.Println("start")
	return nil
}

func (s *Server) Update(sites []dockerproxy.Site) error {
	fmt.Println("update")
	return nil
}

func (s *Server) Name() string {
	return "nginx"
}

type Reloader interface {
	Reload() error
}
