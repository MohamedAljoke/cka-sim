package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/cli/internal/doctor"
)

const usage = `cka-sim — a Kubernetes study environment on your own machine

  cka-sim doctor     check that this machine can run cka-sim
  cka-sim version    print the version
`

// Set by release builds: -ldflags "-X main.version=v0.1.0"
var version = ""

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

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
		return runDoctor(ctx, stdout)
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

func runDoctor(ctx context.Context, stdout io.Writer) error {
	results := doctor.Run(ctx, doctor.Host())
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

func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
