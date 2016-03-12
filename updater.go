package dockerproxy

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
)

func (m *Manager) updater(ctx context.Context) {
	listener := debounceChannel(5*time.Second, m.update)
	for {
		select {
		case _, ok := <-listener:
			if !ok {
				return
			}
			log.Info("updating containers...")
			m.updateContainers()
		}
	}
}

func (m *Manager) updateContainers() error {
	client, err := m.newClient(m.d.Leader, 0)
	if err != nil {
		//todo(nick): build in retries
		return err
	}
	opts := docker.ListContainersOptions{
		Size:    false,
		Filters: map[string][]string{"status": []string{"running"}},
	}

	containers, err := client.ListContainers(opts)
	if err != nil {
		return err
	}
	fmt.Println(containers)
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
