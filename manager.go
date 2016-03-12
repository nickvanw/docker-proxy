package dockerproxy

import (
	"time"

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
	Watchers []string
	Leader   string
	TLS      bool
	Cert     string
	CA       string
	Key      string
}

func New(cfg DockerConfig) (*Manager, error) {
	m := &Manager{
		update: make(chan struct{}),
		d:      cfg,
	}
	// test our leader to make sure our docker connection works
	if _, err := m.newClient(cfg.Leader, 0); err != nil {
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

	log.Info("starting docker event poller(s)")
	go m.watcher(m.d.Leader, ctx)
	for _, v := range m.d.Watchers {
		go m.watcher(v, ctx)
	}

	log.Info("starting polling loop")
	go m.startPoll(ctx, d)
}

func (m *Manager) watcher(addr string, ctx context.Context) {
	tries := 0
	for {
		ll := log.WithField("host", addr)
		client, err := m.newClient(addr, tries)
		tries++
		if err != nil {
			ll.WithError(err).Error("error connecting to Docker")
			continue
		}
		watcher := make(chan *docker.APIEvents)
		if err := client.AddEventListener(watcher); err != nil {
			ll.WithError(err).Error("error starting listener")
			continue
		}
		tries = 0
		ll.Info("polling...")
	Watch:
		for {
			var err error
			select {
			case ev, ok := <-watcher:
				if !ok {
					ll.Info("listener closed")
					break Watch
				}
				if ev == nil {
					continue
				}
				ll := eventLogger(ll, ev)
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
				ll.WithError(err).Error("listener error")
				break Watch
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
func (m *Manager) newClient(addr string, tries int) (*docker.Client, error) {
	backoff(tries)
	log.WithField("addr", addr).Info("creating client")
	client, err := newDockerClient(addr, m.d.TLS, m.d.Cert, m.d.CA, m.d.Key)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(); err != nil {
		return nil, err
	}
	return client, nil
}
