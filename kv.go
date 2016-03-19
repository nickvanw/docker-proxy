package dockerproxy

import (
	"errors"
	"strings"
)

const sep = "="

var (
	// ErrInvalidLine is returned when a malformed environment line is passed
	ErrInvalidLine = errors.New("invalid line passed to parse")
)

func parseKV(kv []string) (map[string]string, error) {
	out := map[string]string{}
	for _, v := range kv {
		index := strings.Index(v, sep)
		if index == -1 {
			return nil, ErrInvalidLine
		}
		data := strings.SplitN(v, sep, 2)
		if len(data) != 2 {
			return nil, ErrInvalidLine
		}
		out[data[0]] = data[1]
	}
	return out, nil
}
