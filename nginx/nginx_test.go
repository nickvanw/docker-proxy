package nginx

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nickvanw/docker-proxy"
)

func TestHostNoSSL(t *testing.T) {
	s, err := New(os.TempDir(), os.TempDir(), "/bin/true")
	if err != nil {
		t.Fatalf("unable to create new server: %s", err)
	}
	host := "www.google.com"
	site := dockerproxy.Site{
		ID:  "my-id-here",
		Env: map[string]string{},
	}
	wantOutput := `
server {
	server_name www.google.com;
	listen 80;
	 
	location / {
		proxy_pass http://my-id-here;
	}
}
`

	buf := bytes.NewBuffer(nil)
	err = s.renderHost(buf, host, site)
	if err != nil {
		t.Fatalf("wanted no error when laying down basic site template, got: %v", err)
	}

	if buf.String() != wantOutput {
		t.Fatalf("want: %q\n got: %q\n", wantOutput, buf.String())
	}
}

func TestHostWithConfigNoSSL(t *testing.T) {
	s, err := New(os.TempDir(), os.TempDir(), "/bin/true")
	if err != nil {
		t.Fatalf("unable to create new server: %s", err)
	}
	host := "www.google.com"
	site := dockerproxy.Site{
		ID:  "my-id-here",
		Env: map[string]string{"NGINX_CLIENT_MAX_BODY_SIZE": "5m"},
	}
	wantOutput := `
server {
	server_name www.google.com;
	listen 80;
	client_max_body_size 5m; 
	location / {
		proxy_pass http://my-id-here;
	}
}
`

	buf := bytes.NewBuffer(nil)
	err = s.renderHost(buf, host, site)
	if err != nil {
		t.Fatalf("wanted no error when laying down basic site template, got: %v", err)
	}

	if buf.String() != wantOutput {
		t.Fatalf("want: %q\n got: %q\n", wantOutput, buf.String())
	}
}

func TestSSLInfo(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ssltest")
	if err != nil {
		t.Fatalf("unable to make temp dir: %q", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("unable to remove tmp files: %s", err)
		}
	}()

	s := &Server{ssl: tmpDir}
	tt := []struct {
		makeFile func()
		host     string
		site     dockerproxy.Site
		wantOut  string
		wantOk   bool
	}{
		{
			makeFile: func() {
				ioutil.WriteFile(filepath.Join(tmpDir, "google.com.crt"), []byte{}, os.FileMode(0400))
				ioutil.WriteFile(filepath.Join(tmpDir, "google.com.key"), []byte{}, os.FileMode(0400))
			},
			host:    "google.com",
			site:    dockerproxy.Site{Env: map[string]string{}},
			wantOut: filepath.Join(tmpDir, "google.com"),
			wantOk:  true,
		},
		{
			makeFile: func() {},
			host:     "www.doesnotmatter.com",
			site:     dockerproxy.Site{Env: map[string]string{}},
			wantOut:  "",
			wantOk:   false,
		},
		{
			makeFile: func() {
				ioutil.WriteFile(filepath.Join(tmpDir, "somekey.crt"), []byte{}, os.FileMode(0400))
				ioutil.WriteFile(filepath.Join(tmpDir, "somekey.key"), []byte{}, os.FileMode(0400))
			},
			host:    "www.doesnotmatch.com",
			site:    dockerproxy.Site{Env: map[string]string{"CERT_NAME": "somekey"}},
			wantOut: filepath.Join(tmpDir, "somekey"),
			wantOk:  true,
		},

		{
			makeFile: func() {
				ioutil.WriteFile(filepath.Join(tmpDir, "almost_matches_both2.crt"), []byte{}, os.FileMode(0400))
				ioutil.WriteFile(filepath.Join(tmpDir, "almost_matches_both.key"), []byte{}, os.FileMode(0400))
			},
			host:    "www.doesnotmatch.com",
			site:    dockerproxy.Site{Env: map[string]string{"CERT_NAME": "almost_matches_both"}},
			wantOk:  false,
			wantOut: "",
		},
	}
	for _, v := range tt {
		v.makeFile()
		got, gotOk := s.sslInfo(v.host, v.site)
		if gotOk != v.wantOk {
			t.Fatalf("wanted ok: %q, got: %q", v.wantOk, gotOk)
		}

		if got != v.wantOut {
			t.Fatalf("wanted: %q, got: %q", v.wantOut, got)
		}
	}
}
func TestHTTPSSite(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ssltest")
	if err != nil {
		t.Fatalf("unable to make temp dir: %q", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("unable to remove temp file: %s", err)
		}
	}()

	wantOut := `
server {
	server_name www.google.com;
	listen 80;
	return 301 https://$host$request_uri;
}

server {
	server_name www.google.com;
	listen 443 ssl;

	ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	ssl_ciphers "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:ECDHE-RSA-DES-CBC3-SHA:ECDHE-ECDSA-DES-CBC3-SHA:EDH-RSA-DES-CBC3-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA";
	ssl_prefer_server_ciphers on;
	ssl_session_timeout 5m;
	ssl_session_cache shared:SSL:50m;

	ssl_certificate ssldir/www.google.com.crt;
	ssl_certificate_key ssldir/www.google.com.key;
	add_header Strict-Transport-Security "max-age=31536000";
	 
	location / {
		proxy_pass http://my-id-upstream;
	}
}
`
	s, err := New(tmpDir, tmpDir, "")
	if err != nil {
		t.Fatalf("unable to create updater: %q", err)
	}

	if err := ioutil.WriteFile(filepath.Join(tmpDir, "www.google.com.crt"), []byte("hi"), os.FileMode(0400)); err != nil {
		t.Fatalf("unable to write crt file: %q", err)
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "www.google.com.key"), []byte("hi"), os.FileMode(0400)); err != nil {
		t.Fatalf("unable to write key file: %q", err)
	}

	buf := bytes.NewBuffer(nil)
	site := dockerproxy.Site{Env: map[string]string{}, ID: "my-id-upstream"}
	err = s.renderHost(buf, "www.google.com", site)
	got := buf.String()
	toCompare := strings.Replace(got, tmpDir, "ssldir", -1)
	if toCompare != wantOut {
		t.Fatalf("want: %q\n got: %q", wantOut, toCompare)
	}
}

func TestHTTPAuthInfo(t *testing.T) {
	dir, err := ioutil.TempDir("", "ssltest")
	if err != nil {
		t.Fatalf("unable to make temp dir: %q", err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("unable to remove temp file: %s", err)
		}
	}()

	tt := []struct {
		site         string
		env          map[string]string
		wantData     []byte
		wantFilename string
		wantErr      error
	}{
		{
			site:         "test.com",
			env:          map[string]string{},
			wantFilename: "",
			wantErr:      nil,
		},
		{
			site:         "test.com",
			env:          map[string]string{"AUTH_USER": "nick", "AUTH_PASS": "password"},
			wantData:     []byte("nick:password"),
			wantFilename: filepath.Join(dir, "test.com.htpasswd"),
			wantErr:      nil,
		},
	}

	for _, v := range tt {
		s := &Server{
			htpasswd: dir,
		}
		site := dockerproxy.Site{
			Env: v.env,
		}
		fn, err := s.httpAuthInfo(v.site, site)

		if err != v.wantErr {
			t.Fatalf("wanted no %v, got: %v", v.wantErr, err)
		}

		if v.wantFilename != fn {
			t.Fatalf("wanted filename: %v, got: %v", v.wantFilename, fn)
		}

		if v.wantFilename != "" {
			fi, err := os.Open(v.wantFilename)
			if err != nil {
				t.Fatalf("error opening wanted file: %v", err)
			}
			data, err := ioutil.ReadAll(fi)
			if err != nil {
				t.Fatalf("unable to fetch data from file: %s", err)
			}

			if !bytes.Equal(data, v.wantData) {
				t.Fatalf("want: %v\n got: %v", string(data), string(v.wantData))
			}

		}
	}
}
