package nginx

import (
	"sync"
	"text/template"

	"github.com/nickvanw/docker-proxy"
)

type Server struct {
	ssl    string
	cfg    string
	reload string

	tpl *template.Template
	sync.Mutex
}

type config struct {
	upstreams map[string]dockerproxy.Mapping
	sites     []site
}

type upstream struct {
	dockerproxy.Mapping
}

type site struct {
	upstream string
	ssl      bool
	ca       string
	key      string
	host     string
}

func New(ssl, config string) (*Server, error) {
	s := &Server{
		ssl:    ssl,
		cfg:    config,
		reload: "nginx -s reload",
	}
	tpl := template.New("nginx")
	t, err := tpl.Parse(nginxTemplate)
	if err != nil {
		return nil, err
	}
	s.tpl = t
	return s, nil
}

func (s *Server) Update(sites []dockerproxy.Site) error {
	s.Lock()
	defer s.Unlock()
	cfg := transform(sites)
	return nil
}

func (s *Server) Name() string {
	return "nginx"
}

func transform(sites []dockerproxy.Sites) config {
	cfg := &config{}
	for _, v := range sites {
		cfg.upstreams[v.ID] = v.Contact
		for _, z := range v.Hosts {
			site := site{upstream: v.ID, host: z}
			cfg.sites = append(cfg.sites, site)
		}
	}
}
