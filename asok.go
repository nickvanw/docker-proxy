package dockerproxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof" // imported for side effects
	"os"
	"sync"

	"github.com/gorilla/mux"
)

type AdminSocket struct {
	sites []Site
	sync.Mutex
}

func (a *AdminSocket) NewSocket() *AdminSocket {
	return &AdminSocket{}
}

func (a *AdminSocket) Name() string {
	return "adminsocket"
}

func (a *AdminSocket) Update(sites []Site) error {
	a.Lock()
	a.sites = sites
	a.Unlock()
	return nil
}

func (a *AdminSocket) Start(asok string) error {
	_ = os.Remove(asok)
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: asok, Net: "unix"})
	if err != nil {
		return err
	}
	r := mux.NewRouter()
	r.HandleFunc("/sites", a.dumpSites)
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	return http.Serve(l, r)
}

func (a *AdminSocket) dumpSites(w http.ResponseWriter, r *http.Request) {
	data := bytes.NewBuffer(nil)
	if err := json.NewEncoder(data).Encode(a.sites); err != nil {
		http.Error(w, "unable to serialize sites", 500)
		return
	}
	io.Copy(w, data)
}
