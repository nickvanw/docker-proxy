package dockerproxy

import (
	"time"
 	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
)

// Manager handles the state and coordinates updates of the mappings
// of Sites -> Containers
type Manager struct {
	update chan struct{}

	d DockerConfig
}

type DockerConfig struct {
	Addr string
	TLS  bool
	Cert string
	CA   string
	Key  string
}

func New(cfg DockerConfig) (*Manager, error) {
	m := &Manager{
		update: make(chan struct{}),
		d:      cfg,
	}
	// test to make sure our docker connection works
	if _, err := m.newClient(); err != nil {
		return nil, err
	}
	return m, nil
}

// Start begins waiting for events on the Docker endpoint, as well as
// polling every `d` seconds
func (m *Manager) Start(ctx context.Context, d time.Duration) {
	// bootstrap initial config

	// start updater
	log.Info("starting update mechanism")
	go m.updater(ctx)

	log.Info("starting docker event poller")
	go m.watcher(ctx)

	log.Info("starting polling loop")
	go m.startPoll(ctx, d)
}

func (m *Manager) watcher(ctx context.Context) {
	for {
		client, err := m.newClient()
		if err != nil {
			continue
		}
		watcher := make(chan *docker.APIEvents)
		if err := client.AddEventListener(watcher); err != nil {
			continue
		}
		log.Info("polling...")
		for {
			var err error
			select {
			case ev := <-watcher:
				fmt.Printf("%#v\n",ev)
				ll := eventLogger(ev)
				ll.Info("received event")
				if isTracked(ev) {
					ll.Info("sending update")
					m.update <- struct{}{}
				}
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
				err = client.Ping()
			}
			if err != nil {
				break
			}
		}
	}
}

// Poll checks the Docker Api every `d` seconds to look for changes
func (m *Manager) startPoll(ctx context.Context, d time.Duration) {
	tkr := time.NewTicker(d)
	defer tkr.Stop()
	for {
		select {
		case <-tkr.C:
			m.update <- struct{}{}
		case <-ctx.Done():
			return
		}
	}
}

// newClient produces a new Docker connection and checks to make sure that it can ping
// the Docker socket. It will properly back off and throttle to avoid too many connections
//todo(nick): use a sync.Pool here?
//todo(nick): add backoff
func (m *Manager) newClient() (*docker.Client, error) {
	client, err := newDockerClient(m.d.Addr, m.d.TLS, m.d.Cert, m.d.CA, m.d.Key)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(); err != nil {
		return nil, err
	}
	return client, nil
}
