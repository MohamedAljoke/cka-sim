package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"

	"golang.org/x/term"

	"github.com/MohamedAljoke/cka-sim/apps/cli/internal/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/cli/internal/doctor"
)

const usage = `cka-sim — a Kubernetes study environment on your own machine

  cka-sim doctor     check that this machine can run cka-sim
  cka-sim up         create the study cluster: 1 control plane, 2 workers
  cka-sim down       delete the study cluster
  cka-sim shell      open a shell on the cluster with kubectl ready
  cka-sim version    print the version
`

// Set by release builds: -ldflags "-X main.version=v0.1.0"
var version = ""

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := run(ctx, os.Args[1:], os.Stdout)
	if err == nil {
		return
	}
	if code, ok := errors.AsType[exitCode](err); ok {
		os.Exit(int(code))
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

// exitCode ends cka-sim with that code and no message, because the command it ran already spoke for itself.
type exitCode int

func (c exitCode) Error() string { return fmt.Sprintf("exit status %d", int(c)) }

func run(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return nil
	}
	command, rest := args[0], args[1:]
	if len(rest) > 0 {
		return fmt.Errorf("%s takes no arguments, got %q", command, rest)
	}
	switch command {
	case "doctor":
		return report(stdout, doctor.Run(ctx, doctor.Host()))
	case "up":
		return runUp(ctx, stdout)
	case "down":
		return runDown(stdout)
	case "shell":
		return runShell()
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "cka-sim %s %s/%s\n", buildVersion(), runtime.GOOS, runtime.GOARCH)
		return nil
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", command, usage)
	}
}

func report(stdout io.Writer, results []doctor.Result) error {
	failed := false
	for _, r := range results {
		line := fmt.Sprintf("%-6s %-18s %s", "["+r.Status+"]", r.Name, r.Detail)
		fmt.Fprintln(stdout, strings.TrimRight(line, " "))
		if r.Fix != "" {
			fmt.Fprintf(stdout, "%-6s %-18s fix: %s\n", "", "", r.Fix)
		}
		failed = failed || r.Status == doctor.Fail
	}
	if failed {
		return errors.New("fix the failed checks above before running cka-sim")
	}
	return nil
}

func problems(results []doctor.Result) []doctor.Result {
	return slices.DeleteFunc(results, func(r doctor.Result) bool { return r.Status == doctor.OK })
}

func runUp(ctx context.Context, stdout io.Writer) error {
	if err := report(stdout, problems(doctor.Run(ctx, doctor.Host()))); err != nil {
		return err
	}
	c, err := cluster.New()
	if err != nil {
		return err
	}
	exists, err := c.Exists()
	if err != nil {
		return err
	}
	if exists {
		fmt.Fprintf(stdout, "cluster %s is already up\n", cluster.Name)
		printHowToConnect(stdout, c)
		return nil
	}

	// kind's Create can't be cancelled, so Ctrl-C returns early and leaves the containers behind.
	created := make(chan error, 1)
	go func() { created <- c.Create(ctx) }()
	select {
	case err := <-created:
		if err != nil {
			return fmt.Errorf("create cluster: %w", err)
		}
	case <-ctx.Done():
		return errors.New("interrupted; run cka-sim down to remove the half-created cluster")
	}
	fmt.Fprintf(stdout, "\ncluster %s is up\n", cluster.Name)
	printHowToConnect(stdout, c)
	return nil
}

func printHowToConnect(stdout io.Writer, c *cluster.Cluster) {
	fmt.Fprintf(stdout, `
  with kubectl on this machine:  kubectl --kubeconfig %q get nodes
  without kubectl:               docker exec -it %s kubectl get nodes
`, c.KubeconfigPath, cluster.ControlPlaneNode)
}

func runDown(stdout io.Writer) error {
	c, err := cluster.New()
	if err != nil {
		return err
	}
	if err := c.Delete(); err != nil {
		return fmt.Errorf("delete cluster: %w", err)
	}
	fmt.Fprintf(stdout, "cluster %s is deleted\n", cluster.Name)
	return nil
}

func runShell() error {
	c, err := cluster.New()
	if err != nil {
		return err
	}
	exists, err := c.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("cluster %s is not up; run cka-sim up first", cluster.Name)
	}

	// Not CommandContext: Ctrl-C is meant for the shell in the container, so it must not kill docker.
	cmd := exec.Command("docker", shellArgs(term.IsTerminal(int(os.Stdin.Fd())))...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err = cmd.Run()
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitCode(exitErr.ExitCode())
	}
	return err
}

// A login shell (-l) reads /etc/profile, so the prompt and PATH match a normal ssh session.
func shellArgs(tty bool) []string {
	args := []string{"exec", "-i"}
	if tty {
		args = append(args, "-t")
	}
	return append(args, cluster.ControlPlaneNode, "bash", "-l")
}

func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
