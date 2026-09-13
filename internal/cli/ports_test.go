package cli

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestImageRepo(t *testing.T) {
	tests := map[string]string{
		"redis:alpine":            "redis",
		"nginx":                   "nginx",
		"ghcr.io/acme/api:latest": "api",
		"postgres@sha256:abcdef":  "postgres",
		"library/redis:7-alpine":  "redis",
	}
	for image, want := range tests {
		if got := imageRepo(image); got != want {
			t.Errorf("imageRepo(%q) = %q, want %q", image, got, want)
		}
	}
}

func TestPidLabel(t *testing.T) {
	tests := []struct {
		name string
		row  portInfo
		want string
	}{
		{"a real process", portInfo{PID: 4182}, "4182"},
		{"no process found", portInfo{PID: 0}, "—"},
		{"docker owns it, even with a stray PID", portInfo{PID: 99, Container: "cache"}, "—"},
	}
	for _, tt := range tests {
		if got := ansi.Strip(pidLabel(tt.row)); got != tt.want {
			t.Errorf("%s: pidLabel(%+v) = %q, want %q", tt.name, tt.row, got, tt.want)
		}
	}
}

func TestWhereLabel(t *testing.T) {
	tests := []struct {
		name string
		row  portInfo
		want string
	}{
		{"docker container", portInfo{Container: "cache-redis"}, "docker: cache-redis"},
		{"project with a branch", portInfo{Dir: "/home/a/api", Branch: "main"}, "/home/a/api (main)"},
		{"project with no branch", portInfo{Dir: "/home/a/api"}, "/home/a/api"},
		{"nothing known", portInfo{}, "—"},
	}
	for _, tt := range tests {
		if got := ansi.Strip(whereLabel(tt.row)); got != tt.want {
			t.Errorf("%s: whereLabel(%+v) = %q, want %q", tt.name, tt.row, got, tt.want)
		}
	}
}
