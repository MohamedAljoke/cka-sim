package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Status string

const (
	OK   Status = "ok"
	Warn Status = "warn"
	Fail Status = "fail"
)

type Result struct {
	Name   string
	Status Status
	Detail string
	Fix    string
}

const (
	MinCPUs             = 2       // kubeadm's minimum for a control-plane node
	MinMemory           = 4 << 30 // 4 GiB, our estimate for a control plane and two workers
	MinInotifyWatches   = 524288  // files watched per user; kind's recommended value
	MinInotifyInstances = 512     // inotify instances per user; kind's recommended value
)

type System struct {
	GOOS     string
	LookPath func(file string) (string, error)
	Run      func(ctx context.Context, name string, args ...string) (stdout []byte, err error)
	ReadFile func(name string) ([]byte, error)
}

func Host() System {
	return System{
		GOOS:     runtime.GOOS,
		LookPath: exec.LookPath,
		Run:      runCommand,
		ReadFile: os.ReadFile,
	}
}

func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, name, args...).Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		err = errors.New(strings.TrimSpace(string(exitErr.Stderr)))
	}
	return out, err
}

type dockerInfo struct {
	ServerVersion   string
	ServerErrors    []string
	OSType          string
	OperatingSystem string
	KernelVersion   string
	NCPU            int
	MemTotal        int64
}

func (i dockerInfo) isDesktop() bool { return strings.Contains(i.OperatingSystem, "Docker Desktop") }
func (i dockerInfo) isWSL() bool {
	return strings.Contains(strings.ToLower(i.KernelVersion), "microsoft")
}

func Run(ctx context.Context, sys System) []Result {
	installed := checkInstalled(sys)
	if installed.Status == Fail {
		return []Result{installed}
	}
	info, running := checkRunning(ctx, sys)
	if running.Status == Fail {
		return []Result{installed, running}
	}
	results := []Result{installed, running, checkLinuxContainers(info), checkCPUs(info), checkMemory(info)}

	// Docker Desktop on Linux runs its own VM, so the host's inotify limits don't reach it.
	// WSL2 distros share one kernel, so there they do.
	hostKernelRunsContainers := !info.isDesktop() || info.isWSL()
	if sys.GOOS == "linux" && hostKernelRunsContainers {
		if inotify, ok := checkInotify(sys); ok {
			results = append(results, inotify)
		}
	}
	return results
}

func checkInstalled(sys System) Result {
	path, err := sys.LookPath("docker")
	if err != nil {
		return Result{
			Name: "docker installed", Status: Fail, Detail: "docker is not on PATH",
			Fix: "install Docker: https://docs.docker.com/get-started/get-docker/",
		}
	}
	return Result{Name: "docker installed", Status: OK, Detail: path}
}

func checkRunning(ctx context.Context, sys System) (dockerInfo, Result) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	out, err := sys.Run(ctx, "docker", "info", "--format", "{{json .}}")

	var info dockerInfo
	// The client still prints its JSON when the daemon is down, with the reason in ServerErrors.
	_ = json.Unmarshal(out, &info)
	if err == nil && info.ServerVersion != "" {
		return info, Result{Name: "docker running", Status: OK, Detail: describeServer(info)}
	}
	reason := firstLine(whyNotRunning(ctx, info, err))
	return info, Result{Name: "docker running", Status: Fail, Detail: reason, Fix: startDockerFix(sys.GOOS, reason)}
}

func describeServer(info dockerInfo) string {
	if info.OperatingSystem == "" {
		return "server " + info.ServerVersion
	}
	return "server " + info.ServerVersion + ", " + info.OperatingSystem
}

func whyNotRunning(ctx context.Context, info dockerInfo, err error) string {
	switch {
	case len(info.ServerErrors) > 0:
		return info.ServerErrors[0]
	case ctx.Err() != nil:
		return "docker info timed out"
	case err != nil:
		return err.Error()
	default:
		return "docker did not answer"
	}
}

func startDockerFix(goos, reason string) string {
	switch {
	case strings.Contains(reason, "permission denied"):
		return "add yourself to the docker group: sudo usermod -aG docker $USER, then log out and back in"
	case goos == "linux":
		return "start Docker: sudo systemctl start docker (or open Docker Desktop if you use it)"
	default:
		return "open Docker Desktop and wait until it says it is running"
	}
}

func checkLinuxContainers(info dockerInfo) Result {
	if info.OSType == "linux" {
		return Result{Name: "linux containers", Status: OK}
	}
	return Result{
		Name: "linux containers", Status: Fail,
		Detail: fmt.Sprintf("docker is running %s containers", info.OSType),
		Fix:    "switch Docker Desktop to Linux containers (tray icon > Switch to Linux containers)",
	}
}

func checkCPUs(info dockerInfo) Result {
	if info.NCPU >= MinCPUs {
		return Result{Name: "cpus", Status: OK, Detail: strconv.Itoa(info.NCPU)}
	}
	return Result{
		Name: "cpus", Status: Warn,
		Detail: fmt.Sprintf("%d available to docker, %d recommended", info.NCPU, MinCPUs),
		Fix:    raiseLimitFix(info, "CPUs", "processors"),
	}
}

func checkMemory(info dockerInfo) Result {
	if info.MemTotal >= MinMemory {
		return Result{Name: "memory", Status: OK, Detail: gib(info.MemTotal)}
	}
	return Result{
		Name: "memory", Status: Warn,
		Detail: fmt.Sprintf("%s available to docker, %s recommended", gib(info.MemTotal), gib(MinMemory)),
		Fix:    raiseLimitFix(info, "memory", "memory"),
	}
}

func raiseLimitFix(info dockerInfo, resource, wslconfigKey string) string {
	switch {
	case info.isDesktop():
		return "give docker more " + resource + ": Docker Desktop > Settings > Resources"
	case info.isWSL():
		return "set " + wslconfigKey + "= under [wsl2] in %UserProfile%\\.wslconfig, then run wsl --shutdown"
	default:
		return "this machine has less " + resource + " than recommended; clusters may be slow or fail to start"
	}
}

func checkInotify(sys System) (Result, bool) {
	watches, err := readInt(sys, "/proc/sys/fs/inotify/max_user_watches")
	if err != nil {
		return Result{}, false
	}
	instances, err := readInt(sys, "/proc/sys/fs/inotify/max_user_instances")
	if err != nil {
		return Result{}, false
	}
	result := Result{
		Name: "inotify limits", Status: OK,
		Detail: fmt.Sprintf("max_user_watches=%d max_user_instances=%d", watches, instances),
	}
	if watches < MinInotifyWatches || instances < MinInotifyInstances {
		result.Status = Warn
		result.Fix = fmt.Sprintf("sudo sysctl fs.inotify.max_user_watches=%d fs.inotify.max_user_instances=%d"+
			" (and put the same values in /etc/sysctl.d/ to keep them after a reboot)",
			MinInotifyWatches, MinInotifyInstances)
	}
	return result, true
}

func readInt(sys System, path string) (int, error) {
	b, err := sys.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func gib(bytes int64) string { return fmt.Sprintf("%.1f GiB", float64(bytes)/(1<<30)) }
