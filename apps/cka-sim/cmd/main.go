package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
)

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

func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
