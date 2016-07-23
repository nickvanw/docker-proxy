package dockerproxy

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

const (
	hostKey = "VIRTUAL_HOST"
	portKey = "VIRTUAL_PORT"
)

// Site represents a single containers' site, which consists
// of the container the site is running in, the IP/Port necessary
// to contact it, as well as 1+ virtual host to advertise.
type Site struct {
	ID      string
	Names   []string
	Contact Mapping
	Hosts   []string
	Env     map[string]string
}

// Mapping represents an IP/Port as well as optional network name
// to map the current container to the container we're representing
type Mapping struct {
	Address string
	Port    int64
	Network string
}

func mapContainers(client *docker.Client, containers []docker.APIContainers) ([]Site, error) {
	list := make([]Site, 0, len(containers))
	myNets, me, err := myInfo(client)
	if err != nil {
		return nil, err
	}

	for _, v := range containers {
		if v.ID == me {
			continue
		}

		info, err := client.InspectContainer(v.ID)
		if err != nil {
			return nil, err
		}
		env, err := parseKV(info.Config.Env)
		if err != nil {
			continue
		}

		hosts, ok := findHosts(env)
		if !ok {
			continue
		}

		portMap := info.NetworkSettings.PortMappingAPI()
		port, _ := findPort(env)
		mapping, ok := findMapping(myNets, v.Networks, portMap, port)
		if !ok {
			continue
		}

		c := Site{
			ID:      v.ID,
			Names:   v.Names,
			Contact: *mapping,
			Hosts:   hosts,
			Env:     env,
		}
		list = append(list, c)
	}
	return list, nil
}

func findHosts(env map[string]string) ([]string, bool) {
	data, ok := env[hostKey]
	if !ok {
		return nil, false
	}
	return strings.Split(data, ","), true
}

func findPort(env map[string]string) (int64, bool) {
	data, ok := env[portKey]
	if !ok {
		return 0, false
	}
	num, err := strconv.Atoi(data)
	return int64(num), err == nil
}

func findMapping(my []string, you docker.NetworkList, ports []docker.APIPort, port int64) (*Mapping, bool) {
	data, ok := findAPIPort(ports, port)
	if !ok {
		return nil, false
	}
	for _, v := range my {
		if net, ok := you.Networks[v]; ok {
			return &Mapping{
				Address: net.IPAddress,
				Network: v,
				Port:    data.PrivatePort,
			}, true
		}
	}
	if data.IP != "" && data.PublicPort != 0 {
		return &Mapping{
			Address: data.IP,
			Network: "public",
			Port:    data.PublicPort,
		}, true
	}
	return nil, false
}

func findAPIPort(ports []docker.APIPort, port int64) (*docker.APIPort, bool) {
	switch len(ports) {
	case 0:
		return nil, false
	case 1:
		return &ports[0], true
	default:
		if port == 0 {
			return nil, false
		}
		for _, v := range ports {
			if v.PrivatePort == port || v.PublicPort == port {
				return &v, true
			}
		}
		return nil, false
	}
}

func myInfo(client *docker.Client) ([]string, string, error) {
	me, err := currentContainerID()
	if err != nil {
		return nil, "", err
	}

	c, err := client.InspectContainer(me)
	if err != nil {
		return nil, me, err
	}
	out := make([]string, 0, len(c.NetworkSettings.Networks))
	networks := c.NetworkSettings.Networks
	for k := range networks {
		out = append(out, k)
	}
	return out, me, nil
}

func currentContainerID() (string, error) {
	if id, ok := os.LookupEnv("CONTAINER_ID"); ok {
		return id, nil
	}

	file, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	regex := "/docker/([[:alnum:]]{64})$"
	re := regexp.MustCompilePOSIX(regex)

	for scanner.Scan() {
		_, lines, err := bufio.ScanLines([]byte(scanner.Text()), true)
		if err == nil {
			if re.MatchString(string(lines)) {
				submatches := re.FindStringSubmatch(string(lines))
				return submatches[1], nil
			}
		}
	}

	return "", errors.New("unable to fetch container ID")
}
