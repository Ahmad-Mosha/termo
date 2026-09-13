package ports

import (
	"reflect"
	"testing"
)

func TestHostPorts(t *testing.T) {
	tests := map[string][]uint32{
		"0.0.0.0:16379->6379/tcp, [::]:16379->6379/tcp": {16379, 16379},
		"0.0.0.0:8080->80/tcp, 0.0.0.0:8443->443/tcp":   {8080, 8443},
		"6379/tcp": nil, // exposed but not published to the host
		"":         nil, // no ports at all, like a stopped container
	}
	for field, want := range tests {
		if got := hostPorts(field); !reflect.DeepEqual(got, want) {
			t.Errorf("hostPorts(%q) = %v, want %v", field, got, want)
		}
	}
}

func TestParseContainers(t *testing.T) {
	// One line per container, exactly as `docker ps --format json` prints it.
	out := []byte(`{"ID":"a1b2c3d4e5f6","Names":"cache-redis","Image":"redis:alpine","Ports":"0.0.0.0:16379->6379/tcp, [::]:16379->6379/tcp","Status":"Up 2 hours"}
{"ID":"f6e5d4c3b2a1","Names":"web,web-alias","Image":"nginx:alpine","Ports":"0.0.0.0:18080->80/tcp","Status":"Up 5 minutes"}
{"ID":"0000deadbeef","Names":"idle","Image":"busybox","Ports":"","Status":"Up 1 second"}
`)
	got := parseContainers(out)

	redis, ok := got[16379]
	if !ok || redis.Name != "cache-redis" || redis.Image != "redis:alpine" || redis.Status != "Up 2 hours" {
		t.Errorf("port 16379 = %+v, ok=%v", redis, ok)
	}
	web, ok := got[18080]
	if !ok || web.Name != "web" || web.ID != "f6e5d4c3b2a1" {
		t.Errorf("port 18080 = %+v, ok=%v; want the first of its comma-separated names", web, ok)
	}
	if len(got) != 2 {
		t.Errorf("got %d ports, want 2 (the idle container publishes none)", len(got))
	}
}

func TestParseContainersIgnoresGarbage(t *testing.T) {
	if got := parseContainers([]byte("not json\n\n{\"ID\":\"x\"}\n")); len(got) != 0 {
		t.Errorf("got %v, want no ports (the one valid line publishes none)", got)
	}
}
