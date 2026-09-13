package cli

import "testing"

func TestParsePort(t *testing.T) {
	for _, s := range []string{"3000", "1", "65535"} {
		if _, err := parsePort(s); err != nil {
			t.Errorf("parsePort(%q) = %v, want no error", s, err)
		}
	}
	for _, s := range []string{"0", "65536", "abc", "-1", ""} {
		if _, err := parsePort(s); err == nil {
			t.Errorf("parsePort(%q) = nil, want an error", s)
		}
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		row  portInfo
		want string
	}{
		{portInfo{Container: "cache-redis", Process: "redis"}, "docker: cache-redis"},
		{portInfo{Process: "node", PID: 123}, "node"},
		{portInfo{PID: 123}, "pid 123"},
	}
	for _, tt := range tests {
		if got := describe(tt.row); got != tt.want {
			t.Errorf("describe(%+v) = %q, want %q", tt.row, got, tt.want)
		}
	}
}
