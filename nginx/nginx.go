package nginx

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/nickvanw/docker-proxy"
)

type Server struct {
	ssl    string
	cfg    string
	reload []string

	tpl  *template.Template
	last []byte

	sync.Mutex
}

func New(ssl, config, reload string) (*Server, error) {
	s := &Server{
		ssl:    ssl,
		cfg:    config,
		reload: strings.Split(reload, " "),
		last:   []byte{},
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
	updater := &Updater{
		Sites:  sites,
		SSLDir: s.ssl,
	}
	data := bytes.NewBuffer(nil)
	if err := s.tpl.Execute(data, updater); err != nil {
		return err
	}
	if !bytes.Equal(data.Bytes(), s.last) {
		fi, err := os.Create(s.cfg)
		if err != nil {
			return err
		}
		_, err = io.Copy(fi, data)
		if err != nil {
			return err
		}
		if err := exec.Command(s.reload[0], s.reload[1:]...).Run(); err != nil {
			return err
		}
		s.last = data.Bytes()
	}
	return nil
}

func (s *Server) Name() string {
	return "nginx"
}

type Updater struct {
	Sites  []dockerproxy.Site
	SSLDir string
}

func (u *Updater) HasSSL(host string) bool {
	pfx := filepath.Join(u.SSLDir, host)
	if _, err := os.Stat(pfx + ".crt"); err != nil {
		return false
	}
	if _, err := os.Stat(pfx + ".key"); err != nil {
		return false
	}
	return true
}

func (u *Updater) SSLPrefix(host string) string {
	pfx := filepath.Join(u.SSLDir, host)
	if _, err := os.Stat(pfx + ".crt"); err != nil {
		return ""
	}
	if _, err := os.Stat(pfx + ".key"); err != nil {
		return ""
	}
	return pfx
}
