package env

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type Container struct {
	Name  string
	State string
}

// Group is one part of the environment and the docker containers behind it.
type Group struct {
	Name       string
	Containers []Container
}

// Groups finds every container the simulator owns: kind labels each node container with its
// cluster's name, and base has a fixed name. Everything is named cka…, so
// docker ps --filter name=cka shows the same set.
func (e *Env) Groups(ctx context.Context) ([]Group, error) {
	var groups []Group
	for _, c := range Clusters {
		list, err := e.containers(ctx, "label=io.x-k8s.kind.cluster="+c.Name)
		if err != nil {
			return nil, err
		}
		groups = append(groups, Group{Name: "cluster " + c.Name, Containers: list})
	}
	list, err := e.containers(ctx, "name=^"+BaseName+"$")
	if err != nil {
		return nil, err
	}
	return append(groups, Group{Name: "base", Containers: list}), nil
}

func (e *Env) containers(ctx context.Context, filter string) ([]Container, error) {
	out, err := e.run(ctx, nil, "docker", "ps", "-a", "--filter", filter, "--format", "{{.Names}}\t{{.State}}")
	if err != nil {
		return nil, err
	}
	var list []Container
	for line := range strings.Lines(out) {
		name, state, ok := strings.Cut(strings.TrimSpace(line), "\t")
		if ok {
			list = append(list, Container{Name: name, State: state})
		}
	}
	return list, nil
}

// Memory sums what these running containers use right now, in bytes.
func (e *Env) Memory(ctx context.Context, names []string) (float64, error) {
	if len(names) == 0 {
		return 0, nil
	}
	args := append([]string{"stats", "--no-stream", "--format", "{{.MemUsage}}"}, names...)
	out, err := e.run(ctx, nil, "docker", args...)
	if err != nil {
		return 0, err
	}
	var total float64
	for line := range strings.Lines(out) {
		// "1.2GiB / 15.5GiB": the part before the slash is what the container uses.
		used, _, _ := strings.Cut(line, "/")
		n, err := parseSize(strings.TrimSpace(used))
		if err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

func parseSize(s string) (float64, error) {
	units := []struct {
		suffix string
		scale  float64
	}{{"GiB", 1 << 30}, {"MiB", 1 << 20}, {"KiB", 1 << 10}, {"GB", 1e9}, {"MB", 1e6}, {"kB", 1e3}, {"B", 1}}
	for _, u := range units {
		if num, ok := strings.CutSuffix(s, u.suffix); ok {
			n, err := strconv.ParseFloat(num, 64)
			return n * u.scale, err
		}
	}
	return 0, fmt.Errorf("unknown size %q", s)
}
