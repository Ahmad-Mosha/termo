package ports

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

// Kill stops whatever is listening on port: the docker container if one
// owns it, or the process. force stops it immediately (SIGKILL, or
// docker kill) instead of asking nicely (SIGTERM, or docker stop).
func Kill(port uint32, force bool) error {
	containers, err := Containers()
	if err != nil {
		return err
	}
	if c, ok := containers[port]; ok {
		return stopContainer(c.ID, force)
	}

	list, err := List()
	if err != nil {
		return err
	}
	for _, p := range list {
		if p.Number != port {
			continue
		}
		if p.PID == 0 {
			return fmt.Errorf("port %d has no process termo can stop", port)
		}
		return killProcess(p.PID, force)
	}
	return fmt.Errorf("nothing is listening on port %d", port)
}

func killProcess(pid int32, force bool) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	if force {
		return proc.Kill()
	}
	return proc.Terminate()
}

func stopContainer(id string, force bool) error {
	verb := "stop"
	if force {
		verb = "kill"
	}
	if out, err := exec.Command("docker", verb, id).CombinedOutput(); err != nil {
		return fmt.Errorf("docker %s: %s", verb, strings.TrimSpace(string(out)))
	}
	return nil
}
