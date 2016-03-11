package dockerproxy

import (
	"time"

	"golang.org/x/net/context"
)

func (m *Manager) updater(ctx context.Context) {
	listener := debounceChannel(10*time.Second, m.update)
	for {
		select {
		case _, ok := <-listener:
			if !ok {
				return
			}

		}
	}
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
