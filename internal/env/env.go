// Package env builds and tears down the exam environment: two kind clusters whose nodes are
// ssh-able exam hosts, and a "base" container you work from that has no Kubernetes tools.
package env

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

const (
	K8sVersion = "v1.35.8"
	NodeImage  = "cka-sim/node:" + K8sVersion
	BaseImage  = "cka-sim/base:latest"
	BaseName   = "cka-base"

	CalicoVersion = "v3.32.2"
	// PodSubnet is Calico's default IP pool, so the manifest needs no editing.
	PodSubnet = "192.168.0.0/16"
)

type Cluster struct {
	Name    string
	Workers int
}

// Clusters mirrors the exam's several clusters: one for everyday workloads, one whose control
// plane and nodes the troubleshooting tasks are allowed to break.
var Clusters = []Cluster{
	{Name: "cka7491", Workers: 2},
	{Name: "cka3962", Workers: 1},
}

// Env runs docker and kind with a private, empty docker config so a broken credential helper
// in the user's own config (common on WSL after Docker Desktop) cannot block public pulls.
type Env struct {
	StateDir string
	Assets   fs.FS
	Out      io.Writer
}

func New(assets fs.FS, out io.Writer) (*Env, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	e := &Env{StateDir: filepath.Join(cache, "cka-sim"), Assets: assets, Out: out}
	if err := os.MkdirAll(e.dockerConfigDir(), 0o755); err != nil {
		return nil, err
	}
	cfg := filepath.Join(e.dockerConfigDir(), "config.json")
	if _, err := os.Stat(cfg); os.IsNotExist(err) {
		if err := os.WriteFile(cfg, []byte("{}\n"), 0o644); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (e *Env) dockerConfigDir() string { return filepath.Join(e.StateDir, "docker") }

// Environ is the process environment every docker, kind, kubectl and task script call runs with.
func (e *Env) Environ() []string {
	return append(os.Environ(), "DOCKER_CONFIG="+e.dockerConfigDir())
}

func (e *Env) command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = e.Environ()
	return cmd
}

func (e *Env) run(ctx context.Context, stdin []byte, name string, args ...string) (string, error) {
	cmd := e.command(ctx, name, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, out.String())
	}
	return out.String(), nil
}

func (e *Env) logf(format string, args ...any) { fmt.Fprintf(e.Out, format+"\n", args...) }

// Unpack writes the embedded assets to disk, because docker build and bash need real files.
func (e *Env) Unpack() (string, error) {
	dir := filepath.Join(e.StateDir, "assets")
	if err := os.RemoveAll(dir); err != nil {
		return "", err
	}
	err := fs.WalkDir(e.Assets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dir, path)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(e.Assets, path)
		if err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(path, ".sh") {
			mode = 0o755
		}
		return os.WriteFile(target, data, mode)
	})
	return dir, err
}

func (e *Env) Up(ctx context.Context) error {
	assets, err := e.Unpack()
	if err != nil {
		return err
	}
	e.logf("• building images (node %s, base) — the first build takes a few minutes", K8sVersion)
	if _, err := e.run(ctx, nil, "docker", "build", "-t", NodeImage, filepath.Join(assets, "images/node")); err != nil {
		return err
	}
	if _, err := e.run(ctx, nil, "docker", "build", "-t", BaseImage, filepath.Join(assets, "images/base")); err != nil {
		return err
	}

	existing, _ := e.run(ctx, nil, "kind", "get", "clusters")
	g, gctx := errgroup.WithContext(ctx)
	for _, c := range Clusters {
		if containsLine(existing, c.Name) {
			e.logf("• cluster %s already exists", c.Name)
			continue
		}
		g.Go(func() error {
			e.logf("• creating cluster %s (1 control plane, %d worker nodes)", c.Name, c.Workers)
			if _, err := e.run(gctx, []byte(kindConfig(c)), "kind", "create", "cluster",
				"--name", c.Name, "--image", NodeImage, "--config", "-"); err != nil {
				return err
			}
			return e.installCNI(gctx, c)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	public, err := e.startBase(ctx)
	if err != nil {
		return err
	}
	if err := e.provisionNodes(ctx, public); err != nil {
		return err
	}
	e.logf("✓ environment ready — run: cka-sim shell")
	return nil
}

// kindConfig swaps kind's own CNI for Calico and keeps the kubelet running on cgroup v1 hosts
// (WSL2 by default): Kubernetes 1.35 refuses cgroup v1 unless failCgroupV1 is turned off.
func kindConfig(c Cluster) string {
	var b strings.Builder
	b.WriteString("kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\n")
	b.WriteString("networking:\n  disableDefaultCNI: true\n  podSubnet: " + PodSubnet + "\n")
	b.WriteString("nodes:\n- role: control-plane\n")
	for range c.Workers {
		b.WriteString("- role: worker\n")
	}
	b.WriteString("kubeadmConfigPatches:\n- |\n  kind: KubeletConfiguration\n  failCgroupV1: false\n")
	return b.String()
}

const noSecurityFS = `{"spec":{"template":{"spec":{
  "volumes":[{"name":"sys-kernel-security","$patch":"delete"}],
  "containers":[{"name":"calico-node","volumeMounts":[{"mountPath":"/sys/kernel/security","$patch":"delete"}]}]}}}}`

// installCNI installs Calico. kind's default CNI enforces NetworkPolicy through nftables queues,
// which the WSL2 kernel lacks (no CONFIG_NFT_QUEUE), so policies would silently allow everything.
// Calico enforces them with iptables, and it is what many exam clusters run.
func (e *Env) installCNI(ctx context.Context, c Cluster) error {
	e.logf("• installing Calico %s on %s", CalicoVersion, c.Name)
	kctx := "kind-" + c.Name
	manifest := "https://raw.githubusercontent.com/projectcalico/calico/" + CalicoVersion + "/manifests/calico.yaml"
	if _, err := e.run(ctx, nil, "kubectl", "--context", kctx, "apply", "--server-side", "-f", manifest); err != nil {
		return err
	}
	// calico-node mounts securityfs only to detect kernel lockdown for its eBPF dataplane; hosts
	// without securityfs (WSL2) cannot create that mount, so the container never starts.
	if _, err := e.run(ctx, nil, "kubectl", "--context", kctx, "-n", "kube-system", "patch", "daemonset", "calico-node",
		"--type=strategic", "-p", noSecurityFS); err != nil {
		return err
	}
	if _, err := e.run(ctx, nil, "kubectl", "--context", kctx, "-n", "kube-system", "rollout", "status", "daemonset/calico-node", "--timeout=10m"); err != nil {
		return err
	}
	_, err := e.run(ctx, nil, "kubectl", "--context", kctx, "wait", "nodes", "--all", "--for=condition=Ready", "--timeout=10m")
	return err
}

func (e *Env) Nodes(ctx context.Context, cluster string) ([]string, error) {
	out, err := e.run(ctx, nil, "kind", "get", "nodes", "--name", cluster)
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}

// provisionNodes gives the candidate user on every node the cluster's admin kubeconfig and
// base's public key — what the exam hosts have ready when you ssh in.
func (e *Env) provisionNodes(ctx context.Context, authorizedKey []byte) error {
	g, gctx := errgroup.WithContext(ctx)
	for _, c := range Clusters {
		g.Go(func() error {
			admin, err := e.run(gctx, nil, "docker", "exec", c.Name+"-control-plane", "cat", "/etc/kubernetes/admin.conf")
			if err != nil {
				return err
			}
			nodes, err := e.Nodes(gctx, c.Name)
			if err != nil {
				return err
			}
			for _, node := range nodes {
				if _, err := e.run(gctx, []byte(admin), "docker", "exec", "-i", node, "bash", "-c",
					"install -d -o candidate -g candidate -m 700 /home/candidate/.kube /home/candidate/.ssh"+
						" && install -o candidate -g candidate -m 600 /dev/stdin /home/candidate/.kube/config"); err != nil {
					return err
				}
				if _, err := e.run(gctx, authorizedKey, "docker", "exec", "-i", node, "install",
					"-o", "candidate", "-g", "candidate", "-m", "600", "/dev/stdin", "/home/candidate/.ssh/authorized_keys"); err != nil {
					return err
				}
			}
			e.logf("• provisioned exam hosts: %s", strings.Join(nodes, ", "))
			return nil
		})
	}
	return g.Wait()
}

// startBase (re)creates the base container; it generates candidate's ssh key itself, the way
// the exam's base host already holds a key every task host trusts.
func (e *Env) startBase(ctx context.Context) (publicKey []byte, err error) {
	_, _ = e.run(ctx, nil, "docker", "rm", "-f", BaseName)
	if _, err := e.run(ctx, nil, "docker", "run", "-d", "--name", BaseName, "--hostname", "base",
		"--network", "kind", BaseImage); err != nil {
		return nil, err
	}
	out, err := e.run(ctx, nil, "docker", "exec", "-u", "candidate", BaseName, "bash", "-c",
		`install -d -m 700 ~/.ssh && ssh-keygen -q -t ed25519 -N "" -C candidate@base -f ~/.ssh/id_ed25519 && cat ~/.ssh/id_ed25519.pub`)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func (e *Env) Down(ctx context.Context) error {
	_, _ = e.run(ctx, nil, "docker", "rm", "-f", BaseName)
	for _, c := range Clusters {
		e.logf("• deleting cluster %s", c.Name)
		if _, err := e.run(ctx, nil, "kind", "delete", "cluster", "--name", c.Name); err != nil {
			return err
		}
	}
	return nil
}

// Shell attaches the terminal to base as candidate — where every exam task starts.
func (e *Env) Shell(ctx context.Context) error {
	cmd := e.command(ctx, "docker", "exec", "-it", "-u", "candidate", "-w", "/home/candidate", BaseName, "bash", "-l")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// Ready reports whether the clusters and base container exist.
func (e *Env) Ready(ctx context.Context) error {
	out, err := e.run(ctx, nil, "kind", "get", "clusters")
	if err != nil {
		return err
	}
	for _, c := range Clusters {
		if !containsLine(out, c.Name) {
			return fmt.Errorf("cluster %s is missing — run: cka-sim up", c.Name)
		}
	}
	if _, err := e.run(ctx, nil, "docker", "inspect", BaseName); err != nil {
		return fmt.Errorf("the base container is missing — run: cka-sim up")
	}
	return nil
}

func containsLine(text, line string) bool {
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}
