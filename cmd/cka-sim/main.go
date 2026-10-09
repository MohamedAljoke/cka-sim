// cka-sim runs a local CKA exam: real clusters, a base host without Kubernetes tools,
// ssh per task, a timed question panel and a grader with partial credit.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	ckasim "github.com/MohamedAljoke/cka-sim"
	"github.com/MohamedAljoke/cka-sim/internal/env"
	"github.com/MohamedAljoke/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/internal/server"
	"github.com/MohamedAljoke/cka-sim/internal/tasks"
	"github.com/MohamedAljoke/cka-sim/internal/terminal"
)

const usage = `cka-sim — a local CKA exam simulator

  cka-sim up                  build images, create the clusters and the base host (once)
  cka-sim shell               a shell on base in this terminal (the panel has one too); ssh <host> per task
  cka-sim exam [-n 16] [-minutes 120] [-port 8080] [-fresh=false] [-resume]
                              rebuild clean clusters, draw an exam, set it up, serve the panel
  cka-sim study [-fresh=false] [-resume] [-port 8080] [id...]
                              no clock: check, reset and read the solution of each task as you go
  cka-sim status              what the simulator runs in docker, its memory, the current session
  cka-sim list                list every task
  cka-sim practice <id>       set up one task, untimed, and print it
  cka-sim check <id>          grade one task now
  cka-sim solution <id>       print the reference solution and explanation (-apply runs it)
  cka-sim selftest [id...]    for each task: setup, expect failure, solve, expect full marks
  cka-sim reset               delete and recreate the whole environment
  cka-sim down                delete the environment
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type app struct {
	env    *env.Env
	tasks  []tasks.Task
	runner grader.Runner
}

func newApp() (*app, error) {
	e, err := env.New(ckasim.Assets, os.Stdout)
	if err != nil {
		return nil, err
	}
	assetsDir, err := e.Unpack()
	if err != nil {
		return nil, err
	}
	all, err := tasks.Load(ckasim.Assets)
	if err != nil {
		return nil, err
	}
	return &app{
		env:    e,
		tasks:  all,
		runner: grader.Runner{AssetsDir: assetsDir, StateDir: e.StateDir, Environ: e.Environ()},
	}, nil
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}
	a, err := newApp()
	if err != nil {
		return err
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "up":
		return a.env.Up(ctx)
	case "down":
		return a.down(ctx)
	case "reset":
		if err := a.down(ctx); err != nil {
			return err
		}
		return a.env.Up(ctx)
	case "shell":
		return a.env.Shell(ctx)
	case "status":
		return a.status(ctx)
	case "list":
		return a.list()
	case "practice":
		return a.withTask(rest, func(t tasks.Task) error { return a.practice(ctx, t) })
	case "check":
		return a.withTask(rest, func(t tasks.Task) error { return a.check(ctx, t) })
	case "solution":
		fset := flag.NewFlagSet("solution", flag.ExitOnError)
		apply := fset.Bool("apply", false, "run the solution against the cluster")
		_ = fset.Parse(rest)
		return a.withTask(fset.Args(), func(t tasks.Task) error { return a.solution(ctx, t, *apply) })
	case "selftest":
		return a.selftest(ctx, rest)
	case "exam":
		return a.exam(ctx, rest)
	case "study":
		return a.study(ctx, rest)
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	}
	return fmt.Errorf("unknown command %q\n\n%s", cmd, usage)
}

func (a *app) withTask(args []string, fn func(tasks.Task) error) error {
	if len(args) != 1 {
		return errors.New("give exactly one task id — see: cka-sim list")
	}
	t, ok := tasks.Find(a.tasks, args[0])
	if !ok {
		return fmt.Errorf("no task %q — see: cka-sim list", args[0])
	}
	return fn(t)
}

// curriculum is the domain order of the CKA curriculum, heaviest first.
var curriculum = []tasks.Domain{tasks.Troubleshooting, tasks.Architecture, tasks.Networking, tasks.Workloads, tasks.Storage}

func (a *app) status(ctx context.Context) error {
	groups, err := a.env.Groups(ctx)
	if err != nil {
		return err
	}
	var running []string
	for _, g := range groups {
		if len(g.Containers) == 0 {
			fmt.Printf("%-17s missing   run: cka-sim up\n", g.Name)
			continue
		}
		names, states := make([]string, len(g.Containers)), map[string]bool{}
		for i, c := range g.Containers {
			names[i] = c.Name
			states[c.State] = true
			if c.State == "running" {
				running = append(running, c.Name)
			}
		}
		state := strings.Join(slices.Sorted(maps.Keys(states)), "/")
		fmt.Printf("%-17s %-9s %s\n", g.Name, state, strings.Join(names, ", "))
	}
	if mem, err := a.env.Memory(ctx, running); err == nil && mem > 0 {
		fmt.Printf("%-17s %.1f GiB\n", "memory", mem/(1<<30))
	}
	fmt.Printf("%-17s %s\n", "session", a.session())
	fmt.Println("\nEvery container is named cka…: docker ps -a --filter name=cka")
	return nil
}

func (a *app) session() string {
	e, err := server.NewStore(a.env.StateDir).Load()
	switch {
	case err != nil:
		return "none"
	case e.EndedAt != nil:
		return fmt.Sprintf("ended · scored %.1f%%", e.Score)
	case e.Study:
		return fmt.Sprintf("study · %d tasks · cka-sim study -resume", len(e.TaskIDs))
	}
	left := int(time.Until(e.Deadline()).Minutes())
	if left <= 0 {
		return "exam · time is up, end it in the panel"
	}
	return fmt.Sprintf("exam · %d tasks · %d min left · cka-sim exam -resume", len(e.TaskIDs), left)
}

func (a *app) list() error {
	for _, d := range curriculum {
		fmt.Printf("\n%s (%d%%)\n", tasks.DomainTitles[d], tasks.DomainWeights[d])
		for _, t := range a.tasks {
			if t.Domain == d {
				fmt.Printf("  %-16s %-40s %2d%%  ssh %s\n", t.ID, t.Title, t.Weight, t.Host)
			}
		}
	}
	return nil
}

func (a *app) setup(ctx context.Context, t tasks.Task) error {
	out, err := a.runner.Script(ctx, t, "setup.sh", 6*time.Minute)
	if err != nil {
		return fmt.Errorf("setting up %s: %w\n%s", t.ID, err, out)
	}
	return nil
}

func (a *app) practice(ctx context.Context, t tasks.Task) error {
	if err := a.env.Ready(ctx); err != nil {
		return err
	}
	fmt.Printf("setting up %s…\n", t.ID)
	if err := a.setup(ctx, t); err != nil {
		return err
	}
	fmt.Printf("\n%s\n\n", render(t))
	fmt.Printf("Work in: cka-sim shell   ·   grade: cka-sim check %s   ·   answer: cka-sim solution %s\n", t.ID, t.ID)
	return nil
}

func render(t tasks.Task) string {
	return fmt.Sprintf("%s  (%d%%)\n%s\n\n  You must connect to the correct host.\n  Failure to do so may result in a zero score.\n\n  [candidate@base] $ ssh %s\n\n%s",
		t.Title, t.Weight, strings.Repeat("─", 60), t.Host, t.Body)
}

func (a *app) check(ctx context.Context, t tasks.Task) error {
	r := a.runner.Grade(ctx, t)
	printResult(t, r)
	return nil
}

func printResult(t tasks.Task, r grader.Result) {
	fmt.Printf("%s — %d/%d points\n", t.ID, r.Earned, r.Total)
	for _, c := range r.Checks {
		mark := "✗"
		if c.Passed {
			mark = "✓"
		}
		fmt.Printf("  %s %s (%d)\n", mark, c.Description, c.Points)
	}
	if r.Error != "" {
		fmt.Printf("  ! %s\n", r.Error)
	}
}

func (a *app) solution(ctx context.Context, t tasks.Task, apply bool) error {
	dir := filepath.Join(a.runner.AssetsDir, t.Dir)
	for _, name := range []string{"explain.md", "solution.sh"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil {
			fmt.Printf("── %s ──\n%s\n", name, data)
		}
	}
	if apply {
		out, err := a.runner.Script(ctx, t, "solution.sh", 5*time.Minute)
		fmt.Print(out)
		return err
	}
	return nil
}

// ordered puts "order: last" tasks after the rest; see tasks.Task.Last.
func ordered(list []tasks.Task) (first, last []tasks.Task) {
	for _, t := range list {
		if t.Last {
			last = append(last, t)
		} else {
			first = append(first, t)
		}
	}
	return first, last
}

// selftest proves every task is fair: right after setup it scores zero, and its reference
// solution earns full marks.
func (a *app) selftest(ctx context.Context, ids []string) error {
	if err := a.env.Ready(ctx); err != nil {
		return err
	}
	selected, err := a.find(ids)
	if err != nil {
		return err
	}
	first, last := ordered(selected)
	failed := 0
	for _, t := range append(first, last...) {
		start := time.Now()
		if err := a.setup(ctx, t); err != nil {
			fmt.Printf("✗ %-16s setup failed: %v\n", t.ID, err)
			failed++
			continue
		}
		before := a.runner.Grade(ctx, t)
		out, err := a.runner.Script(ctx, t, "solution.sh", 5*time.Minute)
		if err != nil {
			fmt.Printf("✗ %-16s solution failed: %v\n%s\n", t.ID, err, out)
			failed++
			continue
		}
		after := a.runner.Grade(ctx, t)
		// An untouched task must score nothing: no free points for doing nothing.
		ok := before.Earned == 0 && after.Total > 0 && after.Earned == after.Total
		mark := "✓"
		if !ok {
			mark = "✗"
			failed++
		}
		fmt.Printf("%s %-16s unsolved %d/%d → solved %d/%d  (%s)\n", mark, t.ID,
			before.Earned, before.Total, after.Earned, after.Total, time.Since(start).Round(time.Second))
		if !ok {
			printResult(t, before)
			printResult(t, after)
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d task(s) failed the selftest", failed)
	}
	return nil
}

// find returns the tasks with these ids, or every task when none are given.
func (a *app) find(ids []string) ([]tasks.Task, error) {
	if len(ids) == 0 {
		all := slices.Clone(a.tasks)
		slices.SortStableFunc(all, func(x, y tasks.Task) int {
			return slices.Index(curriculum, x.Domain) - slices.Index(curriculum, y.Domain)
		})
		return all, nil
	}
	var found []tasks.Task
	for _, id := range ids {
		t, ok := tasks.Find(a.tasks, id)
		if !ok {
			return nil, fmt.Errorf("no task %q — see: cka-sim list", id)
		}
		found = append(found, t)
	}
	return found, nil
}

func (a *app) exam(ctx context.Context, args []string) error {
	fset := flag.NewFlagSet("exam", flag.ExitOnError)
	n := fset.Int("n", 16, "number of tasks")
	port := fset.Int("port", 8080, "port for the exam panel")
	minutes := fset.Int("minutes", 120, "exam duration")
	resume := fset.Bool("resume", false, "serve the current exam again without re-drawing it")
	fresh := fset.Bool("fresh", true, "rebuild the clusters first, so no earlier attempt leaks into this exam")
	_ = fset.Parse(args)

	ln, err := listen(*port)
	if err != nil {
		return err
	}
	defer ln.Close()
	store := server.NewStore(a.env.StateDir)
	if !*resume {
		if err := a.prepare(ctx, *fresh); err != nil {
			return err
		}
		exam := tasks.Draw(a.tasks, *n, rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)))
		if err := a.setupAll(ctx, exam); err != nil {
			return err
		}
		if err := store.Save(&server.Exam{StartedAt: time.Now(), Duration: *minutes * 60, TaskIDs: ids(exam), Flags: map[string]bool{}}); err != nil {
			return err
		}
	}
	return a.serve(ctx, store, ln, "exam ready — the clock is running", "cka-sim exam -resume")
}

// study is the exam panel without a clock: every task (or the ones named) in curriculum order,
// each with its own Check, Reset and Solution buttons.
func (a *app) study(ctx context.Context, args []string) error {
	fset := flag.NewFlagSet("study", flag.ExitOnError)
	port := fset.Int("port", 8080, "port for the study panel")
	resume := fset.Bool("resume", false, "serve the current session again without setting it up")
	fresh := fset.Bool("fresh", true, "rebuild the clusters first, so no earlier attempt gets in the way")
	_ = fset.Parse(args)

	ln, err := listen(*port)
	if err != nil {
		return err
	}
	defer ln.Close()
	store := server.NewStore(a.env.StateDir)
	if !*resume {
		selected, err := a.find(fset.Args())
		if err != nil {
			return err
		}
		if err := a.prepare(ctx, *fresh); err != nil {
			return err
		}
		if err := a.setupAll(ctx, selected); err != nil {
			return err
		}
		if err := store.Save(&server.Exam{Study: true, StartedAt: time.Now(), TaskIDs: ids(selected), Flags: map[string]bool{}}); err != nil {
			return err
		}
	}
	return a.serve(ctx, store, ln, "study session ready — no clock", "cka-sim study -resume")
}

// prepare makes sure the environment exists, rebuilding it first when fresh.
func (a *app) prepare(ctx context.Context, fresh bool) error {
	if fresh {
		fmt.Println("rebuilding a clean environment (about 3 minutes; -fresh=false skips this)…")
		if err := a.down(ctx); err != nil {
			return err
		}
		if err := a.env.Up(ctx); err != nil {
			return err
		}
	}
	return a.env.Ready(ctx)
}

// down deletes the environment and forgets the session and task state that lived in it.
func (a *app) down(ctx context.Context) error {
	if err := a.env.Down(ctx); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(a.env.StateDir, "exam.json")); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.RemoveAll(filepath.Join(a.runner.StateDir, "state"))
}

func ids(list []tasks.Task) []string {
	out := make([]string, len(list))
	for i, t := range list {
		out[i] = t.ID
	}
	return out
}

// listen claims the panel's port before any setup, so a port already in use fails at once
// instead of after minutes of building clusters.
func listen(port int) (net.Listener, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, fmt.Errorf("port %d is taken — is another cka-sim panel running? (use -port to pick another): %w", port, err)
	}
	return ln, nil
}

func (a *app) serve(ctx context.Context, store *server.Store, ln net.Listener, ready, resume string) error {
	ui, err := fs.Sub(ckasim.Assets, "web")
	if err != nil {
		return err
	}
	// The browser terminal's shells outlive a reload, not the panel. Their docker exec clients
	// must not die with ctx: that would leave the shells running in base instead of ending them.
	shells := &terminal.Manager{
		Command: func(tag string) *exec.Cmd { return a.env.ShellCommand(context.WithoutCancel(ctx), tag) },
		Hangup:  func(tag string) { a.endShells(tag) },
	}
	a.endShells("") // left behind by a panel that did not stop cleanly
	srv := &server.Server{Store: store, Tasks: a.tasks, Runner: a.runner, UI: ui, Terminals: shells}
	httpServer := &http.Server{Handler: srv.Handler()}
	go func() {
		<-ctx.Done()
		shells.CloseAll()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdown)
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	fmt.Printf("\n✓ %s\n\n  panel:     http://localhost:%d   — questions and a terminal on base, side by side\n  or a shell in your own terminal:  cka-sim shell\n\nCtrl-C stops the panel and its terminals; %s brings it back.\n", ready, port, resume)
	if err := httpServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (a *app) endShells(tag string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = a.env.EndShells(ctx, tag)
}

func (a *app) setupAll(ctx context.Context, exam []tasks.Task) error {
	first, last := ordered(exam)
	fmt.Printf("setting up %d tasks…\n", len(exam))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(4)
	for _, t := range first {
		g.Go(func() error { return a.setup(gctx, t) })
	}
	if err := g.Wait(); err != nil {
		return err
	}
	for _, t := range last {
		if err := a.setup(ctx, t); err != nil {
			return err
		}
	}
	return nil
}
