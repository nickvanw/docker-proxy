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

// AdminSocket exposes an admin socket for loooking at information about
// the service, as well as what sites are mapped
type AdminSocket struct {
	sites []Site
	sync.Mutex
}

// NewSocket returns a new admin socket for starting
func (a *AdminSocket) NewSocket() *AdminSocket {
	return &AdminSocket{}
}

// Name identifies the admin socket
func (a *AdminSocket) Name() string {
	return "adminsocket"
}

// Update is used to set the current list of sites the admin socket returns
// to be registered with the Manager for automatic updating
func (a *AdminSocket) Update(sites []Site) error {
	a.Lock()
	a.sites = sites
	a.Unlock()
	return nil
}

// Start listens at the specified file for HTTP requests. It is blocking, so
// it should be run in a goroutine
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
