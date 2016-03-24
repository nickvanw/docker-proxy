package dockerproxy

import (
	"reflect"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func TestHosts(t *testing.T) {
	tt := []struct {
		m    map[string]string
		want []string
		ok   bool
	}{
		{
			m:    map[string]string{"VIRTUAL_HOST": "www.google.com"},
			want: []string{"www.google.com"},
			ok:   true,
		},
		{
			m:    map[string]string{},
			want: nil,
			ok:   false,
		},
		{
			m:    map[string]string{"VIRTUAL_HOST": "www.google.com,google.com,yahoo.com"},
			want: []string{"www.google.com", "google.com", "yahoo.com"},
			ok:   true,
		},
	}
	for _, v := range tt {
		hosts, ok := findHosts(v.m)
		if ok != v.ok {
			t.Errorf("wanted ok: %v, got: %v", v.ok, ok)
		}
		if !reflect.DeepEqual(v.want, hosts) {
			t.Fatalf("wanted hosts: %#v, got: %#v", v.want, hosts)
		}
	}
}

func TestPorts(t *testing.T) {
	tt := []struct {
		m    map[string]string
		want int64
		ok   bool
	}{
		{
			m:    map[string]string{"VIRTUAL_PORT": "9001"},
			want: 9001,
			ok:   true,
		},
		{
			m:    map[string]string{},
			want: 0,
			ok:   false,
		},
		{
			m:    map[string]string{"VIRTUAL_PORT": "ENOTASTRING"},
			want: 0,
			ok:   false,
		},
	}
	for _, v := range tt {
		port, ok := findPort(v.m)
		if ok != v.ok {
			t.Errorf("wanted ok: %v, got: %v", v.ok, ok)
		}
		if port != v.want {
			t.Fatalf("wanted port: %d, got: %d", v.want, port)
		}
	}
}

func TestNetworkMapping(t *testing.T) {
	tt := []struct {
		my    []string
		you   docker.NetworkList
		ports []docker.APIPort
		port  int64
		want  *Mapping
		ok    bool
	}{
		{
			my:    []string{},
			you:   docker.NetworkList{Networks: map[string]docker.ContainerNetwork{}},
			ports: []docker.APIPort{{PrivatePort: 9001, PublicPort: 80, IP: "4.2.2.1"}},
			port:  80,
			want:  &Mapping{Address: "4.2.2.1", Port: 80, Network: "public"},
			ok:    true,
		},
		{
			my:    []string{},
			you:   docker.NetworkList{Networks: map[string]docker.ContainerNetwork{}},
			ports: []docker.APIPort{{PrivatePort: 9001, PublicPort: 80, IP: "4.2.2.1"}, {PrivatePort: 9002, PublicPort: 81, IP: "4.2.2.1"}},
			port:  81,
			want:  &Mapping{Address: "4.2.2.1", Port: 81, Network: "public"},
			ok:    true,
		},
		{
			my:    []string{},
			you:   docker.NetworkList{Networks: map[string]docker.ContainerNetwork{}},
			ports: []docker.APIPort{{PrivatePort: 9001, PublicPort: 80, IP: "4.2.2.1"}, {PrivatePort: 9002, PublicPort: 81, IP: "4.2.2.1"}},
			port:  82,
			want:  nil,
			ok:    false,
		},
		{
			my:    []string{"overlay-net"},
			you:   docker.NetworkList{Networks: map[string]docker.ContainerNetwork{"overlay-net": {IPAddress: "10.0.1.2"}}},
			ports: []docker.APIPort{{PrivatePort: 9001}},
			port:  0,
			want:  &Mapping{Address: "10.0.1.2", Port: 9001, Network: "overlay-net"},
			ok:    true,
		},
		{
			my:    []string{},
			you:   docker.NetworkList{Networks: map[string]docker.ContainerNetwork{}},
			ports: []docker.APIPort{},
			port:  0,
			want:  nil,
			ok:    false,
		},
		{
			my:    []string{},
			you:   docker.NetworkList{Networks: map[string]docker.ContainerNetwork{}},
			ports: []docker.APIPort{{PrivatePort: 9001, PublicPort: 80, IP: "4.2.2.1"}, {PrivatePort: 9002, PublicPort: 81, IP: "4.2.2.1"}},
			port:  0,
			want:  nil,
			ok:    false,
		},
	}
	for _, v := range tt {
		mapping, ok := findMapping(v.my, v.you, v.ports, v.port)
		if ok != v.ok {
			t.Errorf("wanted ok: %v, got: %v", v.ok, ok)
		}
		if !reflect.DeepEqual(mapping, v.want) {
			t.Fatalf("wanted mapping: %#v, got: %#v", v.want, mapping)
		}
	}
}
