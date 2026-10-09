// cka-sim runs a local CKA exam: real clusters, a base host without Kubernetes tools,
// ssh per task, a timed question panel and a grader with partial credit.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	ckasim "github.com/MohamedAljoke/cka-sim"
	"github.com/MohamedAljoke/cka-sim/internal/env"
	"github.com/MohamedAljoke/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/internal/server"
	"github.com/MohamedAljoke/cka-sim/internal/tasks"
)

const usage = `cka-sim — a local CKA exam simulator

  cka-sim up                  build images, create the clusters and the base host (once)
  cka-sim shell               open a shell on base — where you work; ssh <host> per task
  cka-sim exam [-n 16] [-minutes 120] [-port 8080] [-fresh=false] [-resume]
                              rebuild clean clusters, draw an exam, set it up, serve the panel
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
		return a.env.Down(ctx)
	case "reset":
		if err := a.env.Down(ctx); err != nil {
			return err
		}
		return a.env.Up(ctx)
	case "shell":
		return a.env.Shell(ctx)
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

func (a *app) list() error {
	for _, d := range []tasks.Domain{tasks.Troubleshooting, tasks.Architecture, tasks.Networking, tasks.Workloads, tasks.Storage} {
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
	selected := a.tasks
	if len(ids) > 0 {
		selected = nil
		for _, id := range ids {
			t, ok := tasks.Find(a.tasks, id)
			if !ok {
				return fmt.Errorf("no task %q", id)
			}
			selected = append(selected, t)
		}
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

func (a *app) exam(ctx context.Context, args []string) error {
	fset := flag.NewFlagSet("exam", flag.ExitOnError)
	n := fset.Int("n", 16, "number of tasks")
	port := fset.Int("port", 8080, "port for the exam panel")
	minutes := fset.Int("minutes", 120, "exam duration")
	resume := fset.Bool("resume", false, "serve the current exam again without re-drawing it")
	fresh := fset.Bool("fresh", true, "rebuild the clusters first, so no earlier attempt leaks into this exam")
	_ = fset.Parse(args)

	store := server.NewStore(a.env.StateDir)
	if !*resume {
		if *fresh {
			fmt.Println("rebuilding a clean environment (about 3 minutes; -fresh=false skips this)…")
			if err := a.env.Down(ctx); err != nil {
				return err
			}
			if err := a.env.Up(ctx); err != nil {
				return err
			}
		}
		if err := a.env.Ready(ctx); err != nil {
			return err
		}
		exam := tasks.Draw(a.tasks, *n, rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)))
		if err := a.setupAll(ctx, exam); err != nil {
			return err
		}
		ids := make([]string, len(exam))
		for i, t := range exam {
			ids[i] = t.ID
		}
		if err := store.Save(&server.Exam{StartedAt: time.Now(), Duration: *minutes * 60, TaskIDs: ids, Flags: map[string]bool{}}); err != nil {
			return err
		}
	}

	ui, err := fs.Sub(ckasim.Assets, "web")
	if err != nil {
		return err
	}
	srv := &server.Server{Store: store, Tasks: a.tasks, Runner: a.runner, UI: ui}
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	httpServer := &http.Server{Addr: addr, Handler: srv.Handler()}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdown)
	}()
	fmt.Printf("\n✓ exam ready — the clock is running\n\n  exam panel:  http://localhost:%d\n  your shell:  cka-sim shell   (in another terminal)\n\nCtrl-C stops the panel; cka-sim exam -resume brings it back.\n", *port)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
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
