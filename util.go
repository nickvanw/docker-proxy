package dockerproxy

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

var trackedEvents = []string{"create", "destroy", "start", "stop", "die", "restart", "connect", "disconnect"}

func isTracked(ev *docker.APIEvents) bool {
	for _, v := range trackedEvents {
		if ev.Action == v {
			return true
		}
	}
	return false
}

func eventLogger(ev *docker.APIEvents) *log.Entry {
	ll := log.WithFields(log.Fields{
		"action": ev.Action,
		"type":   ev.Type,
	})
	return ll.WithFields(mapToLog(ev.Actor.Attributes))
}
func mapToLog(m map[string]string) log.Fields {
	fields := log.Fields{}
	for k, v := range m {
		fields[k] = v
	}
	return fields
}
