package doctor

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

func TestHealthyLinuxPassesEveryCheck(t *testing.T) {
	results := Run(context.Background(), healthyLinux().System())

	if len(results) != 7 {
		t.Fatalf("got %d results, want 7: %+v", len(results), results)
	}
	for _, r := range results {
		if r.Status != OK {
			t.Errorf("%s: got %s (%s), want ok", r.Name, r.Status, r.Detail)
		}
	}
}

func TestDockerMissingStopsEarly(t *testing.T) {
	sys := healthyLinux()
	sys.dockerOnPath = false

	results := Run(context.Background(), sys.System())

	if len(results) != 1 || results[0].Status != Fail {
		t.Fatalf("got %+v, want a single failure", results)
	}
}

func TestDockerNotRunning(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		infoJSON  string
		infoErr   error
		wantInFix string
	}{
		{
			name:      "daemon stopped on linux",
			goos:      "linux",
			infoJSON:  `{"ServerErrors":["Cannot connect to the Docker daemon at unix:///var/run/docker.sock"]}`,
			infoErr:   errors.New("exit status 1"),
			wantInFix: "systemctl start docker",
		},
		{
			name:      "no permission on the socket",
			goos:      "linux",
			infoErr:   errors.New("permission denied while trying to connect to the Docker daemon socket"),
			wantInFix: "usermod -aG docker",
		},
		{
			name:      "docker desktop closed on macOS",
			goos:      "darwin",
			infoErr:   errors.New("Cannot connect to the Docker daemon"),
			wantInFix: "open Docker Desktop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := healthyLinux()
			sys.goos, sys.infoJSON, sys.infoErr = tt.goos, tt.infoJSON, tt.infoErr

			results := Run(context.Background(), sys.System())

			running := mustFind(t, results, "docker running")
			if running.Status != Fail || len(results) != 2 {
				t.Fatalf("got %+v, want to stop after a failed docker running check", results)
			}
			if !strings.Contains(running.Fix, tt.wantInFix) {
				t.Errorf("fix %q does not mention %q", running.Fix, tt.wantInFix)
			}
		})
	}
}

func TestWindowsContainersFail(t *testing.T) {
	sys := healthyLinux()
	sys.goos = "windows"
	sys.infoJSON = dockerInfoJSON("windows", "Docker Desktop", "10.0.22631", 8, 16<<30)

	results := Run(context.Background(), sys.System())

	if r := mustFind(t, results, "linux containers"); r.Status != Fail {
		t.Errorf("got %s, want fail", r.Status)
	}
}

func TestCgroupV1Warns(t *testing.T) {
	tests := []struct {
		name      string
		kernel    string
		wantInFix string
	}{
		{"WSL2", "5.15.167.4-microsoft-standard-WSL2", "cgroup_no_v1=all"},
		{"plain linux", "6.8.0", "systemd.unified_cgroup_hierarchy=1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := healthyLinux()
			sys.infoJSON = strings.Replace(dockerInfoJSON("linux", "Ubuntu 24.04", tt.kernel, 8, 16<<30),
				`"CgroupVersion":"2"`, `"CgroupVersion":"1"`, 1)

			r := mustFind(t, Run(context.Background(), sys.System()), "cgroup version")

			if r.Status != Warn || !strings.Contains(r.Fix, tt.wantInFix) {
				t.Errorf("got %+v, want a warning whose fix mentions %q", r, tt.wantInFix)
			}
		})
	}
}

func TestLowResourcesPointToWhereTheLimitIsSet(t *testing.T) {
	tests := []struct {
		name            string
		operatingSystem string
		kernel          string
		wantInFix       string
	}{
		{"docker desktop", "Docker Desktop", "6.10.14-linuxkit", "Docker Desktop > Settings > Resources"},
		{"docker engine inside WSL2", "Ubuntu 24.04", "5.15.167.4-microsoft-standard-WSL2", ".wslconfig"},
		{"plain linux", "Ubuntu 24.04", "6.8.0", "this machine has less"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := healthyLinux()
			sys.infoJSON = dockerInfoJSON("linux", tt.operatingSystem, tt.kernel, 1, 2<<30)

			results := Run(context.Background(), sys.System())

			for _, name := range []string{"cpus", "memory"} {
				r := mustFind(t, results, name)
				if r.Status != Warn {
					t.Errorf("%s: got %s, want warn", name, r.Status)
				}
				if !strings.Contains(r.Fix, tt.wantInFix) {
					t.Errorf("%s: fix %q does not mention %q", name, r.Fix, tt.wantInFix)
				}
			}
		})
	}
}

func TestInotify(t *testing.T) {
	t.Run("ubuntu defaults are too low", func(t *testing.T) {
		sys := healthyLinux()
		sys.procFiles["/proc/sys/fs/inotify/max_user_watches"] = "8192"
		sys.procFiles["/proc/sys/fs/inotify/max_user_instances"] = "128"

		r := mustFind(t, Run(context.Background(), sys.System()), "inotify limits")

		if r.Status != Warn || !strings.Contains(r.Fix, "sysctl") {
			t.Errorf("got %+v, want a warning with a sysctl fix", r)
		}
	})

	t.Run("checked for docker desktop inside WSL2", func(t *testing.T) {
		sys := healthyLinux()
		sys.infoJSON = dockerInfoJSON("linux", "Docker Desktop", "5.15.167.4-microsoft-standard-WSL2", 8, 16<<30)

		if _, ok := find(Run(context.Background(), sys.System()), "inotify limits"); !ok {
			t.Error("want an inotify result")
		}
	})

	skipped := map[string]func(*fakeSystem){
		"skipped when /proc can't be read": func(s *fakeSystem) { s.procFiles = nil },
		"skipped on macOS":                 func(s *fakeSystem) { s.goos = "darwin" },
		"skipped for docker desktop's own VM on linux": func(s *fakeSystem) {
			s.infoJSON = dockerInfoJSON("linux", "Docker Desktop", "6.10.14-linuxkit", 8, 16<<30)
		},
	}
	for name, change := range skipped {
		t.Run(name, func(t *testing.T) {
			sys := healthyLinux()
			change(&sys)

			if r, ok := find(Run(context.Background(), sys.System()), "inotify limits"); ok {
				t.Errorf("got %+v, want no inotify result", r)
			}
		})
	}
}

type fakeSystem struct {
	goos         string
	dockerOnPath bool
	infoJSON     string
	infoErr      error
	procFiles    map[string]string
}

func (f fakeSystem) System() System {
	return System{
		GOOS: f.goos,
		LookPath: func(string) (string, error) {
			if !f.dockerOnPath {
				return "", errors.New("not found")
			}
			return "/usr/bin/docker", nil
		},
		Run: func(context.Context, string, ...string) ([]byte, error) {
			return []byte(f.infoJSON), f.infoErr
		},
		ReadFile: func(name string) ([]byte, error) {
			content, ok := f.procFiles[name]
			if !ok {
				return nil, fs.ErrNotExist
			}
			return []byte(content), nil
		},
	}
}

func dockerInfoJSON(osType, operatingSystem, kernel string, cpus int, memory int64) string {
	return fmt.Sprintf(`{"ServerVersion":"28.3.0","OSType":%q,"OperatingSystem":%q,"KernelVersion":%q,"CgroupVersion":"2","NCPU":%d,"MemTotal":%d}`,
		osType, operatingSystem, kernel, cpus, memory)
}

func healthyLinux() fakeSystem {
	return fakeSystem{
		goos:         "linux",
		dockerOnPath: true,
		infoJSON:     dockerInfoJSON("linux", "Ubuntu 24.04", "6.8.0", 8, 16<<30),
		procFiles: map[string]string{
			"/proc/sys/fs/inotify/max_user_watches":   "524288\n",
			"/proc/sys/fs/inotify/max_user_instances": "512\n",
		},
	}
}

func find(results []Result, name string) (Result, bool) {
	for _, r := range results {
		if r.Name == name {
			return r, true
		}
	}
	return Result{}, false
}

func mustFind(t *testing.T, results []Result, name string) Result {
	t.Helper()
	r, ok := find(results, name)
	if !ok {
		t.Fatalf("no %q result in %+v", name, results)
	}
	return r
}
