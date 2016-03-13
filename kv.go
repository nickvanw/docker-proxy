package dockerproxy

import (
	"errors"
	"strings"
)

const sep = "="

var (
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
		out[data[0]] = data[1]
	}
	return out, nil
}
