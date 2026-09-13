package ports

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// Container is the docker container behind a published port.
type Container struct {
	ID     string
	Name   string
	Image  string
	Status string // docker's own words, like "Up 2 hours"
}

// Containers maps each host port docker has published to the container
// behind it. It returns an empty map, with no error, when docker isn't
// installed or its daemon isn't running: termo works fine without docker.
func Containers() (map[uint32]Container, error) {
	out, err := exec.Command("docker", "ps", "--format", "json").Output()
	if err != nil {
		return map[uint32]Container{}, nil
	}
	return parseContainers(out), nil
}

// dockerPS is the subset of `docker ps --format json`'s fields we use.
type dockerPS struct {
	ID     string `json:"ID"`
	Names  string `json:"Names"`
	Image  string `json:"Image"`
	Ports  string `json:"Ports"`
	Status string `json:"Status"`
}

// parseContainers reads docker ps's newline-delimited JSON.
func parseContainers(out []byte) map[uint32]Container {
	m := make(map[uint32]Container)
	for line := range bytes.Lines(out) {
		var d dockerPS
		if json.Unmarshal(bytes.TrimSpace(line), &d) != nil {
			continue
		}
		c := Container{ID: d.ID, Name: strings.Split(d.Names, ",")[0], Image: d.Image, Status: d.Status}
		for _, port := range hostPorts(d.Ports) {
			m[port] = c
		}
	}
	return m
}

// hostPorts reads the host-side ports from a Ports field such as
// "0.0.0.0:16379->6379/tcp, [::]:16379->6379/tcp". A container port that
// isn't published to the host (no "->") is skipped.
func hostPorts(field string) []uint32 {
	var out []uint32
	for _, mapping := range strings.Split(field, ",") {
		host, _, ok := strings.Cut(strings.TrimSpace(mapping), "->")
		if !ok {
			continue
		}
		if i := strings.LastIndex(host, ":"); i != -1 {
			if port, err := strconv.ParseUint(host[i+1:], 10, 32); err == nil {
				out = append(out, uint32(port))
			}
		}
	}
	return out
}
