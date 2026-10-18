package usecases

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	taskusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

var (
	t0      = time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	catalog = []tasks.Task{
		breaking(task("tr-etcd", tasks.Troubleshooting, 7, tasks.Last)),
		task("ar-rbac", tasks.Architecture, 4, ""),
		task("nw-svc", tasks.Networking, 3, ""),
	}
)

func TestBeginSetsUpEveryDrawnTaskThenStartsTheClock(t *testing.T) {
	f := newFixture(t, nil)

	got, err := f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	if err != nil || len(got.Tasks) != 3 || !got.SettingUp() {
		t.Fatalf("Begin = %+v, %v", got, err)
	}
	f.wait()

	e := f.get()
	if !e.Started.Equal(t0) || !e.Deadline.Equal(t0.Add(30*time.Minute)) {
		t.Errorf("started %v, deadline %v", e.Started, e.Deadline)
	}
	for _, task := range e.Tasks {
		if task.Setup != exams.SetUp {
			t.Errorf("%s: setup %s %q", task.ID, task.Setup, task.SetupError)
		}
	}
	if saved := f.repo.saved(); saved == nil || saved.Started.IsZero() {
		t.Errorf("saved %+v", saved)
	}
}

func TestBeginSetsUpTheLastTaskAfterTheOthers(t *testing.T) {
	f := newFixture(t, nil)

	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()

	if ran := f.runner.ran(); len(ran) != 3 || ran[2] != "setup tr-etcd" {
		t.Errorf("ran %v", ran)
	}
}

func TestAFailedSetupMarksItsTaskAndKeepsWhy(t *testing.T) {
	f := newFixture(t, nil)
	f.runner.fail("setup nw-svc", labs.ScriptResult{ExitCode: 1, Output: "error: no such CRD\n"})
	f.runner.err("setup ar-rbac", errors.New("connection refused"))

	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()

	e := f.get()
	svc, rbac := e.Tasks[e.Index("nw-svc")], e.Tasks[e.Index("ar-rbac")]
	if svc.Setup != exams.SetupFailed || !strings.Contains(svc.SetupError, "code 1") || !strings.Contains(svc.SetupError, "no such CRD") {
		t.Errorf("nw-svc: %+v", svc)
	}
	if rbac.Setup != exams.SetupFailed || !strings.Contains(rbac.SetupError, "connection refused") {
		t.Errorf("ar-rbac: %+v", rbac)
	}
}

func TestBeginHealsWhatPracticeBrokeBeforeTheSetups(t *testing.T) {
	f := newFixture(t, nil)
	f.tasks.SetUp.Execute(context.Background(), owner, "tr-etcd")

	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()

	if ran := f.runner.ran(); len(ran) != 5 || ran[1] != "reset tr-etcd" || ran[4] != "setup tr-etcd" {
		t.Errorf("ran %v", ran)
	}
}

func TestBeginRefusesBadLimits(t *testing.T) {
	f := newFixture(t, nil)

	if _, err := f.uc.Begin.Execute(context.Background(), owner, 0, 30); !errors.Is(err, exams.ErrInvalid) {
		t.Errorf("got %v", err)
	}
}

func TestBeginNeedsAReadyLab(t *testing.T) {
	f := newFixture(t, nil)
	f.lab.lab = labs.Lab{State: labs.Starting}

	if _, err := f.uc.Begin.Execute(context.Background(), owner, 3, 30); !errors.Is(err, labs.ErrNotReady) || f.repo.saved() != nil {
		t.Errorf("got %v", err)
	}
}

func TestBeginIsUnavailableWhenTheLabDoesNotAnswer(t *testing.T) {
	f := newFixture(t, nil)
	f.lab.err = errors.New("dial tcp: connection refused")

	if _, err := f.uc.Begin.Execute(context.Background(), owner, 3, 30); !errors.Is(err, labs.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestBeginIsUnavailableWhenTheExamCannotBeSaved(t *testing.T) {
	f := newFixture(t, nil)
	f.repo.saveErr = errors.New("disk full")

	_, err := f.uc.Begin.Execute(context.Background(), owner, 3, 30)

	if !errors.Is(err, exams.ErrUnavailable) || len(f.runner.ran()) != 0 {
		t.Errorf("got %v, ran %v", err, f.runner.ran())
	}
	if _, err := f.uc.Get.Execute(context.Background(), owner); !errors.Is(err, exams.ErrNoExam) {
		t.Errorf("an unsaved exam shows: %v", err)
	}
}

func TestOnlyOneExamRunsAtATime(t *testing.T) {
	f := newFixture(t, nil)
	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()

	if _, err := f.uc.Begin.Execute(context.Background(), owner, 3, 30); !errors.Is(err, exams.ErrRunning) {
		t.Errorf("got %v", err)
	}
}

func TestEndGradesEveryPlayableTaskThenScoresTheExam(t *testing.T) {
	f := newFixture(t, nil)
	f.runner.fail("setup nw-svc", labs.ScriptResult{ExitCode: 1})
	f.runner.fail("check ar-rbac", labs.ScriptResult{Output: "PASS 2 role\nFAIL 2 binding\n"})
	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()
	f.clock = t0.Add(10 * time.Minute)

	got, err := f.uc.End.Execute(context.Background(), owner)
	if err != nil || !got.Ended.Equal(t0.Add(10*time.Minute)) {
		t.Fatalf("End = %+v, %v", got, err)
	}
	f.wait()

	e := f.get()
	rbac := e.Tasks[e.Index("ar-rbac")]
	if e.Scoring() || rbac.Grade == nil || rbac.Grade.Earned != 2 || rbac.Grade.Total != 4 {
		t.Errorf("ar-rbac graded %+v, scoring %v", rbac.Grade, e.Scoring())
	}
	if slices.Contains(f.runner.ran(), "check nw-svc") || e.Tasks[e.Index("nw-svc")].Grade != nil {
		t.Errorf("checked a task whose setup failed: %v", f.runner.ran())
	}
}

func TestAScoredExamTidiesEveryTask(t *testing.T) {
	f := newFixture(t, nil)
	f.runner.fail("setup nw-svc", labs.ScriptResult{ExitCode: 1})
	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()

	f.uc.End.Execute(context.Background(), owner)
	f.wait()

	ran := f.runner.ran()
	tidied := []string{taskusecases.TidyScript("ar-rbac"), taskusecases.TidyScript("nw-svc"), taskusecases.TidyScript("tr-etcd")}
	if len(ran) < 4 || ran[len(ran)-4] != "reset tr-etcd" || !slices.Equal(slices.Sorted(slices.Values(ran[len(ran)-3:])), tidied) {
		t.Errorf("ran %v", ran)
	}
}

func TestABrokenCheckScoresZeroAndSaysWhy(t *testing.T) {
	f := newFixture(t, nil)
	f.runner.fail("check ar-rbac", labs.ScriptResult{ExitCode: 127, Output: "kubectl: not found\n"})
	f.runner.err("check nw-svc", errors.New("connection refused"))
	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()

	f.uc.End.Execute(context.Background(), owner)
	f.wait()

	e := f.get()
	rbac, svc := e.Tasks[e.Index("ar-rbac")], e.Tasks[e.Index("nw-svc")]
	if rbac.Grade == nil || rbac.Grade.Total != 0 || !strings.Contains(rbac.CheckError, "code 127") {
		t.Errorf("ar-rbac: %+v %q", rbac.Grade, rbac.CheckError)
	}
	if svc.Grade == nil || svc.Grade.Total != 0 || !strings.Contains(svc.CheckError, "connection refused") {
		t.Errorf("nw-svc: %+v %q", svc.Grade, svc.CheckError)
	}
}

func TestAnExamCannotEndWhileSettingUpOrTwice(t *testing.T) {
	f := newFixture(t, &exams.Exam{Minutes: 30, Tasks: []exams.Task{{ID: "nw-svc", Setup: exams.Preparing}}})

	if _, err := f.uc.End.Execute(context.Background(), owner); !errors.Is(err, exams.ErrSettingUp) {
		t.Errorf("setting up: got %v", err)
	}

	f = newFixture(t, &exams.Exam{Started: t0, Deadline: t0.Add(time.Hour), Ended: t0, Scored: t0})
	if _, err := f.uc.End.Execute(context.Background(), owner); !errors.Is(err, exams.ErrEnded) {
		t.Errorf("ended: got %v", err)
	}
}

func TestReadingAnExamPastItsDeadlineEndsItAtTheDeadline(t *testing.T) {
	f := newFixture(t, nil)
	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()
	f.clock = t0.Add(45 * time.Minute)

	got, err := f.uc.Get.Execute(context.Background(), owner)
	if err != nil || !got.Ended.Equal(t0.Add(30*time.Minute)) {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	f.wait()

	if e := f.get(); e.Scoring() {
		t.Errorf("not scored: %+v", e)
	}
}

func TestFlagMarksATaskAndSavesIt(t *testing.T) {
	f := newFixture(t, nil)
	f.uc.Begin.Execute(context.Background(), owner, 3, 30)
	f.wait()

	if err := f.uc.Flag.Execute(context.Background(), owner, "nw-svc", true); err != nil {
		t.Fatal(err)
	}

	saved := f.repo.saved()
	if !saved.Tasks[saved.Index("nw-svc")].Flagged || !f.get().Tasks[saved.Index("nw-svc")].Flagged {
		t.Errorf("saved %+v", saved.Tasks)
	}
	if err := f.uc.Flag.Execute(context.Background(), owner, "wl-scale", true); !errors.Is(err, exams.ErrNotInExam) {
		t.Errorf("unknown task: got %v", err)
	}
}

func TestAnEndedExamCannotBeFlagged(t *testing.T) {
	f := newFixture(t, &exams.Exam{Tasks: []exams.Task{{ID: "nw-svc"}}, Started: t0, Ended: t0, Scored: t0})

	if err := f.uc.Flag.Execute(context.Background(), owner, "nw-svc", true); !errors.Is(err, exams.ErrEnded) {
		t.Errorf("got %v", err)
	}
}

func TestNoExamIsNotFound(t *testing.T) {
	f := newFixture(t, nil)

	if _, err := f.uc.Get.Execute(context.Background(), owner); !errors.Is(err, exams.ErrNoExam) {
		t.Errorf("Get: %v", err)
	}
	if err := f.uc.Flag.Execute(context.Background(), owner, "nw-svc", true); !errors.Is(err, exams.ErrNoExam) {
		t.Errorf("Flag: %v", err)
	}
	if _, err := f.uc.End.Execute(context.Background(), owner); !errors.Is(err, exams.ErrNoExam) {
		t.Errorf("End: %v", err)
	}
	if err := f.uc.Discard.Execute(context.Background(), owner); !errors.Is(err, exams.ErrNoExam) {
		t.Errorf("Discard: %v", err)
	}
}

func TestOnlyAScoredExamIsDiscarded(t *testing.T) {
	f := newFixture(t, &exams.Exam{Started: t0, Deadline: t0.Add(time.Hour)})
	if err := f.uc.Discard.Execute(context.Background(), owner); !errors.Is(err, exams.ErrRunning) {
		t.Errorf("running: got %v", err)
	}

	f = newFixture(t, &exams.Exam{Started: t0, Ended: t0, Scored: t0})
	if err := f.uc.Discard.Execute(context.Background(), owner); err != nil || f.repo.saved() != nil {
		t.Errorf("scored: got %v, saved %+v", err, f.repo.saved())
	}
	if _, err := f.uc.Get.Execute(context.Background(), owner); !errors.Is(err, exams.ErrNoExam) {
		t.Errorf("after discard: %v", err)
	}
}

func TestResumeFinishesTheSetupsAStoppedServerLeft(t *testing.T) {
	f := newFixture(t, &exams.Exam{Minutes: 30, Tasks: []exams.Task{
		{ID: "ar-rbac", Setup: exams.SetUp},
		{ID: "nw-svc", Setup: exams.Preparing},
	}})

	f.uc.Resume.Execute(context.Background()) //nolint:errcheck // the fake repository never fails here
	f.wait()

	e := f.get()
	if ran := f.runner.ran(); !slices.Equal(ran, []string{"setup nw-svc"}) || e.SettingUp() {
		t.Errorf("ran %v, exam %+v", ran, e)
	}
}

func TestResumeFinishesTheChecksAStoppedServerLeft(t *testing.T) {
	f := newFixture(t, &exams.Exam{Started: t0, Ended: t0, Tasks: []exams.Task{
		{ID: "ar-rbac", Setup: exams.SetUp, Grade: &tasks.Grade{Earned: 1, Total: 1}},
		{ID: "nw-svc", Setup: exams.SetUp},
	}})

	f.uc.Resume.Execute(context.Background()) //nolint:errcheck // the fake repository never fails here
	f.wait()

	checks := slices.DeleteFunc(f.runner.ran(), func(s string) bool { return !strings.HasPrefix(s, "check ") })
	if !slices.Equal(checks, []string{"check nw-svc"}) || f.get().Scoring() {
		t.Errorf("ran %v", checks)
	}
}

func TestAnUnreadableSavedExamIsUnavailable(t *testing.T) {
	repo := &fakeExamRepository{loadErr: errors.New("unexpected end of JSON input")}
	uc := New(repo, taskusecases.UseCases{}, nil, time.Now, rand.New(rand.NewPCG(1, 2)))

	if err := uc.Resume.Execute(context.Background()); !errors.Is(err, exams.ErrUnavailable) {
		t.Errorf("Resume: got %v, want ErrUnavailable", err)
	}
	if _, err := uc.Get.Execute(context.Background(), owner); !errors.Is(err, exams.ErrUnavailable) {
		t.Errorf("Get: got %v, want ErrUnavailable", err)
	}
}

type fixture struct {
	t      *testing.T
	uc     UseCases
	tasks  taskusecases.UseCases
	repo   *fakeExamRepository
	runner *fakeRunner
	lab    *fakeLab
	clock  time.Time
}

func newFixture(t *testing.T, saved *exams.Exam) *fixture {
	t.Helper()
	f := &fixture{t: t, repo: &fakeExamRepository{exam: saved}, runner: newFakeRunner(), lab: &fakeLab{lab: labs.Lab{State: labs.Ready}}, clock: t0}
	f.tasks = taskusecases.New(fakeTaskRepository{}, f.runner)
	f.uc = New(f.repo, f.tasks, labusecases.New(f.lab).Get, func() time.Time { return f.clock }, rand.New(rand.NewPCG(1, 2)))
	return f
}

func (f *fixture) wait() {
	f.t.Helper()
	s, err := f.uc.Get.ss.of(owner)
	if err != nil {
		f.t.Fatal(err)
	}
	s.work.Wait()
}

func (f *fixture) get() exams.Exam {
	f.t.Helper()
	e, err := f.uc.Get.Execute(context.Background(), owner)
	if err != nil {
		f.t.Fatal(err)
	}
	return e
}

func task(id string, domain tasks.Domain, weight int, order tasks.Order) tasks.Task {
	return tasks.Task{ID: id, Domain: domain, Weight: weight, Order: order, Host: "cka-sim-control-plane", Setup: "setup " + id, Check: "check " + id}
}

func breaking(t tasks.Task) tasks.Task {
	t.Reset = "reset " + t.ID
	return t
}

type fakeTaskRepository struct{}

func (fakeTaskRepository) List(context.Context) ([]tasks.Task, error) { return catalog, nil }

func (fakeTaskRepository) Get(_ context.Context, id string) (tasks.Task, error) {
	i := slices.IndexFunc(catalog, func(t tasks.Task) bool { return t.ID == id })
	if i < 0 {
		return tasks.Task{}, tasks.ErrNotFound
	}
	return catalog[i], nil
}

type fakeRunner struct {
	mu      sync.Mutex
	scripts []string
	results map[string]labs.ScriptResult
	errs    map[string]error
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{results: map[string]labs.ScriptResult{}, errs: map[string]error{}}
}

func (r *fakeRunner) fail(script string, result labs.ScriptResult) { r.results[script] = result }

func (r *fakeRunner) err(script string, err error) { r.errs[script] = err }

func (r *fakeRunner) RunScript(_ context.Context, _, _, script string) (labs.ScriptResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scripts = append(r.scripts, script)
	if result, ok := r.results[script]; ok {
		return result, nil
	}
	if strings.HasPrefix(script, "check ") {
		return labs.ScriptResult{Output: "PASS 1 done\n"}, r.errs[script]
	}
	return labs.ScriptResult{}, r.errs[script]
}

func (r *fakeRunner) ran() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.scripts)
}

// owner is the student every exam test acts for.
const owner = "alice"

// fakeExamRepository holds the one owner's exam.
type fakeExamRepository struct {
	mu      sync.Mutex
	exam    *exams.Exam
	loadErr error
	saveErr error
}

func (r *fakeExamRepository) Load(string) (*exams.Exam, error) {
	if r.exam == nil {
		return nil, r.loadErr
	}
	e := r.exam.Clone()
	return &e, r.loadErr
}

func (r *fakeExamRepository) Save(_ string, e exams.Exam) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	e = e.Clone()
	r.exam = &e
	return nil
}

func (r *fakeExamRepository) Owners() ([]string, error) {
	if r.exam == nil && r.loadErr == nil {
		return nil, nil
	}
	return []string{owner}, nil
}

func (r *fakeExamRepository) Clear(string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.exam = nil
	return nil
}

func (r *fakeExamRepository) saved() *exams.Exam {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.exam
}

type fakeLab struct {
	labusecases.LabProvider
	lab labs.Lab
	err error
}

func (l *fakeLab) Get(context.Context, string) (labs.Lab, error) { return l.lab, l.err }
