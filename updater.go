package dockerproxy

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
)

var containerListOptions = docker.ListContainersOptions{
	Size:    false,
	Filters: map[string][]string{"status": []string{"running"}},
}

func (m *Manager) updater(ctx context.Context) {
	listener := debounceChannel(3*time.Second, m.update)
	for {
		select {
		case _, ok := <-listener:
			if !ok {
				return
			}
			t := time.Now()
			log.Info("updating containers")
			if err := m.updateContainers(); err != nil {
				log.WithError(err).Error("unable to update container mappings")
			}
			log.WithField("duration", time.Since(t)).Info("updated containers")
		case <-ctx.Done():
			return
		}
	}
}

func (m *Manager) updateContainers() error {
	client, err := m.newClient(m.d.Leader, 0)
	if err != nil {
		return err
	}

	containers, err := client.ListContainers(containerListOptions)
	if err != nil {
		return err
	}

	mapping, err := mapContainers(client, containers)
	if err != nil {
		return err
	}

	for _, v := range m.notify {
		if err := v.Update(mapping); err != nil {
			log.WithError(err).WithField("notifier", v.Name()).Error("unable to update mapping")
		}
	}

	return nil
}

// debounceChannel takes an input channel which may be noisy, and makes sure that the returned
// channel only gets called after `interval` time without being called.
func debounceChannel(interval time.Duration, input chan struct{}) chan struct{} {
	output := make(chan struct{})
	go func() {
		for {
			_, ok := <-input
			if !ok {
				close(output)
				return
			}
			select {
			case <-input:
			case <-time.After(interval):
				output <- struct{}{}
			}
		}
	}()
	return output
}
