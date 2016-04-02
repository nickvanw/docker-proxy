package nginx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/nickvanw/docker-proxy"
)

const (
	sslNameEnv   = "CERT_NAME"
	httpAuthUser = "AUTH_USER"
	httpAuthPass = "AUTH_PASS"
)

// Server represents an nginx server to configure
type Server struct {
	Syslog string

	ssl      string
	htpasswd string
	cfg      string
	reload   []string

	headerTpl   *template.Template
	upstreamTpl *template.Template
	noSSLTpl    *template.Template
	sslTpl      *template.Template

	last []byte

	sync.Mutex
}

// New returns a new instance of nginx to configure
func New(ssl, config, reload string) (*Server, error) {
	s := &Server{
		ssl:    ssl,
		cfg:    config,
		reload: strings.Split(reload, " "),
	}

	s.headerTpl = template.Must(template.New("nginxHeader").Parse(nginxHeader))
	s.upstreamTpl = template.Must(template.New("nginxUpstream").Parse(nginxUpstream))
	s.noSSLTpl = template.Must(template.New("nginxNoSSL").Parse(nginxNoSSL))
	s.sslTpl = template.Must(template.New("nginxWithSSL").Parse(nginxWithSSL))
	template.Must(s.noSSLTpl.Parse(nginxOptions))
	template.Must(s.sslTpl.Parse(nginxOptions))

	return s, nil
}

// Name returns the string representation of the server
func (s *Server) Name() string {
	return "nginx"
}

// Update is called with a list of sites to update the nginx configuration
func (s *Server) Update(sites []dockerproxy.Site) error {
	s.Lock()
	defer s.Unlock()

	var data bytes.Buffer
	if err := s.headerTpl.Execute(&data, s); err != nil {
		return err
	}

	for _, v := range sites {
		if err := s.renderSite(&data, v); err != nil {
			return err
		}
	}

	if !bytes.Equal(data.Bytes(), s.last) {
		fi, err := os.Create(s.cfg)
		if err != nil {
			return err
		}
		_, err = io.Copy(fi, &data)
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

func (s *Server) renderSite(wr io.Writer, site dockerproxy.Site) error {
	if err := s.upstreamTpl.Execute(wr, site); err != nil {
		return err
	}

	for _, h := range site.Hosts {
		if err := s.renderHost(wr, h, site); err != nil {
			return err
		}
	}
	return nil
}

type dockersite struct {
	Host      string
	ID        string
	SSLPrefix string
	Config    map[string]string
}

func (s *Server) renderHost(wr io.Writer, host string, site dockerproxy.Site) error {
	d := dockersite{
		Host:   host,
		ID:     site.ID,
		Config: envToDirectives(site.Env),
	}
	var ok bool

	passwd, err := s.httpAuthInfo(host, site)
	if err != nil {
		return err
	}
	if passwd != "" {
		d.Config["auth_basic"] = `"closed site"`
		d.Config["auth_basic_user_file"] = passwd
	}
	if d.SSLPrefix, ok = s.sslInfo(host, site); ok {
		if err := s.sslTpl.Execute(wr, d); err != nil {
			return err
		}
	} else {
		if err := s.noSSLTpl.Execute(wr, d); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) httpAuthInfo(host string, site dockerproxy.Site) (string, error) {
	var user, pass string
	var ok bool

	out := bytes.NewBuffer(nil)
	if user, ok = site.Env[httpAuthUser]; ok {
		out.Write([]byte(user + ":"))
	} else {
		return "", nil
	}

	if pass, ok = site.Env[httpAuthPass]; ok {
		out.Write([]byte(pass))
	} else {
		return "", nil
	}
	fn := fmt.Sprintf("%s.htpasswd", host)
	p := filepath.Join(s.htpasswd, fn)
	fi, err := os.Create(p)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(fi, out)
	return p, err
}

func (s *Server) sslInfo(host string, site dockerproxy.Site) (string, bool) {
	key := host
	if name, ok := site.Env[sslNameEnv]; ok {
		key = name
	}
	pfx := filepath.Join(s.ssl, key)
	if _, err := os.Stat(pfx + ".crt"); err != nil {
		return "", false
	}
	if _, err := os.Stat(pfx + ".key"); err != nil {
		return "", false
	}
	return pfx, true
}
