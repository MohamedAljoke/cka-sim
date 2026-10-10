package exam

import (
	"context"
	"errors"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/grader"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

func TestBeginSetsEveryTaskUpThenStartsTheClock(t *testing.T) {
	s, store, r := newSession(t, nil)

	released := make(chan struct{})
	if err := s.Begin(s.cfg.Catalog, 2, 30, func() { close(released) }); err != nil {
		t.Fatal(err)
	}
	first, _ := s.Current()
	<-released
	e := waitStarted(t, s)

	if len(first.Tasks) != 2 || first.Prepared() {
		t.Errorf("right after Begin: %+v, want 2 tasks and no clock yet", first)
	}
	for _, task := range e.Tasks {
		if task.Setup != Ready {
			t.Errorf("%s: setup %q, want ready", task.ID, task.Setup)
		}
	}
	if !e.Started.Equal(now) || !e.Deadline.Equal(now.Add(30*time.Minute)) {
		t.Errorf("started %v deadline %v, want %v and 30 minutes later", e.Started, e.Deadline, now)
	}
	if got := r.ran("setup.sh"); !slices.Equal(got, []string{"tr-svc", "wl-scale"}) {
		t.Errorf("setup ran for %v", got)
	}
	if saved := store.exam; saved == nil || !saved.Prepared() {
		t.Errorf("saved %+v, want the prepared exam", saved)
	}
}

func TestBeginMarksAFailedSetup(t *testing.T) {
	s, _, r := newSession(t, nil)
	r.setupErr = map[string]error{"wl-scale": errors.New("namespace stuck")}

	e := begin(t, s, 2)

	task := e.Tasks[slices.IndexFunc(e.Tasks, func(t Task) bool { return t.ID == "wl-scale" })]
	if task.Setup != Failed || !strings.Contains(task.SetupError, "namespace stuck") {
		t.Errorf("got %+v, want a failed setup with the error", task)
	}
	if !e.Prepared() {
		t.Error("a failed setup must not stop the clock from starting")
	}
}

func TestBeginSetsUpLastTasksAfterTheRest(t *testing.T) {
	s, _, r := openSession(t, withLast(), nil)
	r.hold = make(chan struct{})

	if err := s.Begin(s.cfg.Catalog, 3, 30, func() {}); err != nil {
		t.Fatal(err)
	}
	for len(r.ran("setup.sh")) < 2 {
		time.Sleep(time.Millisecond)
	}
	if got := r.ran("setup.sh"); slices.Contains(got, "ar-etcd") {
		t.Errorf("ar-etcd set up alongside %v, want it after them", got)
	}
	close(r.hold)
	waitSetUp(t, s)

	if got := r.ran("setup.sh"); len(got) != 3 || !slices.Contains(got, "ar-etcd") {
		t.Errorf("setup ran for %v, want all three", got)
	}
}

func TestTheClockStartsBeforeTheLastTasks(t *testing.T) {
	s, _, r := openSession(t, withLast(), nil)
	r.holdLast()
	released := make(chan struct{})

	if err := s.Begin(s.cfg.Catalog, 3, 30, func() { close(released) }); err != nil {
		t.Fatal(err)
	}
	e := waitStarted(t, s)

	if i := slices.IndexFunc(e.Tasks, func(t Task) bool { return t.ID == "ar-etcd" }); e.Tasks[i].Setup != Preparing {
		t.Errorf("ar-etcd is %q when the clock starts, want it still preparing", e.Tasks[i].Setup)
	}
	select {
	case <-released:
		t.Fatal("released while ar-etcd is still setting up")
	default:
	}
	r.releaseLast()
	<-released
	if e := waitSetUp(t, s); e.SettingUp() {
		t.Errorf("got %+v, want every task set up", e)
	}
}

func TestTheLastTasksTimeIsGivenBack(t *testing.T) {
	s, _, r := openSession(t, withLast(), nil)
	clock := struct {
		sync.Mutex
		now time.Time
	}{now: now}
	s.cfg.Now = func() time.Time { clock.Lock(); defer clock.Unlock(); return clock.now }
	r.holdLast()

	if err := s.Begin(s.cfg.Catalog, 3, 30, func() {}); err != nil {
		t.Fatal(err)
	}
	waitStarted(t, s)
	for !slices.Contains(r.ran("setup.sh"), "ar-etcd") {
		time.Sleep(time.Millisecond)
	}
	clock.Lock()
	clock.now = now.Add(5 * time.Minute)
	clock.Unlock()
	r.releaseLast()
	e := waitSetUp(t, s)

	if !e.Started.Equal(now) || !e.Deadline.Equal(now.Add(35*time.Minute)) {
		t.Errorf("started %v deadline %v, want %v and 30+5 minutes later", e.Started, e.Deadline, now)
	}
}

func TestBeginHealsBeforeAnySetup(t *testing.T) {
	s, _, r := openSession(t, withReset("tr-svc", "wl-scale"), nil)

	begin(t, s, 1)

	if got := r.order(); len(got) != 3 || got[0] != "tr-svc/reset.sh" || got[1] != "wl-scale/reset.sh" {
		t.Errorf("ran %v, want every reset.sh in the catalog before the setup", got)
	}
}

func TestBeginWaitsForTheLabBeforeAnySetup(t *testing.T) {
	s, _, r := openSession(t, withReset("tr-svc"), nil)
	lab := make(chan struct{})
	s.cfg.Ready = func(context.Context) error { <-lab; return nil }

	if err := s.Begin(s.cfg.Catalog, 2, 30, func() {}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if got := r.order(); len(got) != 0 {
		t.Fatalf("ran %v before the lab was ready", got)
	}
	close(lab)
	waitStarted(t, s)
	if got := r.ran("setup.sh"); len(got) != 2 {
		t.Errorf("setup ran for %v, want both tasks", got)
	}
}

func TestBeginFailsEveryTaskWhenTheLabFails(t *testing.T) {
	s, _, r := newSession(t, nil)
	s.cfg.Ready = func(context.Context) error { return errors.New("cluster is not up") }

	e := begin(t, s, 2)

	for _, task := range e.Tasks {
		if task.Setup != Failed || !strings.Contains(task.SetupError, "cluster is not up") {
			t.Errorf("got %+v, want a failed setup with the lab's error", task)
		}
	}
	if got := r.order(); len(got) != 0 {
		t.Errorf("ran %v without a lab", got)
	}
}

func TestBeginRefusesASecondExam(t *testing.T) {
	s, _, _ := newSession(t, nil)
	begin(t, s, 1)

	if err := s.Begin(s.cfg.Catalog, 1, 30, func() {}); !errors.Is(err, ErrExists) {
		t.Errorf("got %v, want ErrExists", err)
	}
}

func TestEndGradesEveryTask(t *testing.T) {
	s, store, r := newSession(t, nil)
	r.checkOut = map[string]string{"wl-scale": "PASS 2 scaled\nPASS 2 ready\n", "tr-svc": "FAIL 3 endpoints\n"}
	begin(t, s, 2)

	end(t, s)
	e := waitScored(t, s)

	if !e.Over() || store.exam.Scored.IsZero() {
		t.Error("the exam is not over, or not saved as scored")
	}
	scores := map[string][2]int{}
	for _, task := range e.Tasks {
		scores[task.ID] = [2]int{task.Result.Earned, task.Result.Total}
	}
	if want := map[string][2]int{"wl-scale": {4, 4}, "tr-svc": {0, 3}}; !maps.Equal(scores, want) {
		t.Errorf("earned/total %v, want %v", scores, want)
	}
	if percent, _ := Score(e, s.Weights()); percent != 40 {
		t.Errorf("score %d%%, want 40%% (wl-scale weighs 4 of 10)", percent)
	}
}

func TestEndTidiesAfterChecks(t *testing.T) {
	s, _, r := openSession(t, withReset("tr-svc"), nil)
	begin(t, s, 2)
	r.forget()

	end(t, s)
	waitScored(t, s)

	got := r.order()
	if len(got) == 5 {
		slices.Sort(got[3:])
	}
	if want := []string{"tr-svc/reset.sh", "tr-svc/tidy", "wl-scale/tidy"}; len(got) != 5 || !slices.Equal(got[2:], want) {
		t.Errorf("ran %v, want both checks then %v", got, want)
	}
}

func TestEndRecordsACrashedCheck(t *testing.T) {
	s, _, r := newSession(t, nil)
	begin(t, s, 1)
	r.checkErr = errors.New("kubectl: not found")

	end(t, s)
	e := waitScored(t, s)

	if task := e.Tasks[0]; task.Result == nil || task.Result.Total != 0 || !strings.Contains(task.CheckError, "kubectl: not found") {
		t.Errorf("got %+v, want 0 points and the check's error", task)
	}
}

func TestEndRefuses(t *testing.T) {
	t.Run("no exam", func(t *testing.T) {
		s, _, _ := newSession(t, nil)
		if _, err := s.End(func() {}); !errors.Is(err, ErrNoExam) {
			t.Errorf("got %v", err)
		}
	})
	t.Run("still preparing", func(t *testing.T) {
		s, _, _ := newSession(t, &Exam{Minutes: 30, Tasks: []Task{{ID: "wl-scale", Setup: Preparing}}})
		if _, err := s.End(func() {}); err == nil || !strings.Contains(err.Error(), "still being set up") {
			t.Errorf("got %v", err)
		}
	})
	t.Run("a task still setting up", func(t *testing.T) {
		saved := &Exam{Minutes: 30, Started: now, Deadline: now.Add(30 * time.Minute),
			Tasks: []Task{{ID: "wl-scale", Setup: Ready}, {ID: "tr-svc", Setup: Preparing}}}
		s, _, _ := newSession(t, saved)
		if _, err := s.End(func() {}); err == nil || !strings.Contains(err.Error(), "still being set up") {
			t.Errorf("got %v", err)
		}
	})
	t.Run("twice", func(t *testing.T) {
		s, _, _ := newSession(t, nil)
		begin(t, s, 1)
		end(t, s)
		if _, err := s.End(func() {}); err == nil || !strings.Contains(err.Error(), "already ended") {
			t.Errorf("got %v", err)
		}
	})
}

func TestBeginDrawsOnlyFromThePool(t *testing.T) {
	s, _, _ := newSession(t, nil)
	pool := tasks.Filter(s.cfg.Catalog, []tasks.Domain{tasks.Workloads}, nil)

	if err := s.Begin(pool, 5, 30, func() {}); err != nil {
		t.Fatal(err)
	}

	if e, _ := s.Current(); len(e.Tasks) != 1 || e.Tasks[0].ID != "wl-scale" {
		t.Errorf("tasks = %+v, want only wl-scale", e.Tasks)
	}
	if err := s.Begin(nil, 1, 30, func() {}); err == nil {
		t.Error("an empty pool began an exam")
	}
}

func TestBeginReplacesAScoredExam(t *testing.T) {
	s, store, _ := newSession(t, nil)
	begin(t, s, 1)
	end(t, s)
	waitScored(t, s)

	if err := s.Begin(s.cfg.Catalog, 2, 30, func() {}); err != nil {
		t.Fatal(err)
	}

	if e, _ := s.Current(); len(e.Tasks) != 2 || e.Over() || store.saved().Over() {
		t.Errorf("got %+v, want a new exam of 2 tasks in place of the old one", e)
	}
}

func TestEndReturnsBeforeTheChecksFinish(t *testing.T) {
	s, _, r := newSession(t, nil)
	r.checkOut = map[string]string{"wl-scale": "PASS 4 scaled\n"}
	begin(t, s, 2)
	r.holdChecks()

	released := make(chan struct{})
	e, err := s.End(func() { close(released) })

	if err != nil {
		t.Fatal(err)
	}
	if !e.Over() || !e.Scored.IsZero() || e.Tasks[0].Result != nil {
		t.Errorf("End returned %+v, want ended, not scored, no results yet", e)
	}
	if err := s.Begin(s.cfg.Catalog, 1, 30, func() {}); !errors.Is(err, ErrScoring) {
		t.Errorf("Begin while scoring = %v, want ErrScoring", err)
	}
	select {
	case <-released:
		t.Fatal("released before the checks came back")
	default:
	}
	r.releaseChecks()
	<-released
	if got := waitScored(t, s); got.Tasks[0].Result == nil || got.Tasks[1].Result == nil {
		t.Errorf("got %+v, want every task graded", got)
	}
}

func TestResumeGradesAnEndedExam(t *testing.T) {
	saved := &Exam{Minutes: 30, Started: now, Deadline: now.Add(30 * time.Minute), Ended: now,
		Tasks: []Task{{ID: "wl-scale", Setup: Ready, Result: &grader.Result{Earned: 4, Total: 4}}, {ID: "tr-svc", Setup: Ready}}}
	s, _, r := newSession(t, saved)

	if !s.Resume(func() {}) {
		t.Fatal("Resume had nothing to do")
	}
	e := waitScored(t, s)

	if got := r.ran("check.sh"); !slices.Equal(got, []string{"tr-svc"}) {
		t.Errorf("checks ran for %v, want only the ungraded tr-svc", got)
	}
	if e.Tasks[1].Result == nil {
		t.Error("tr-svc is still ungraded")
	}
}

func TestFlag(t *testing.T) {
	s, store, _ := newSession(t, nil)
	begin(t, s, 2)

	if err := s.Flag("wl-scale", true); err != nil {
		t.Fatal(err)
	}

	if i := slices.IndexFunc(store.exam.Tasks, func(t Task) bool { return t.ID == "wl-scale" }); !store.exam.Tasks[i].Flagged {
		t.Error("the flag was not saved")
	}
	if err := s.Flag("nope", true); err == nil || !strings.Contains(err.Error(), `"nope" is not in this exam`) {
		t.Errorf("flagging an unknown task: %v", err)
	}
}

func TestResumeSetsUpOnlyWhatWasInterrupted(t *testing.T) {
	saved := &Exam{Minutes: 30, Tasks: []Task{{ID: "tr-svc", Setup: Ready}, {ID: "wl-scale", Setup: Preparing}}}
	s, _, r := newSession(t, saved)

	if !s.Resume(func() {}) {
		t.Fatal("Resume had nothing to do")
	}
	e := waitStarted(t, s)

	if got := r.ran("setup.sh"); !slices.Equal(got, []string{"wl-scale"}) {
		t.Errorf("setup ran for %v, want only wl-scale", got)
	}
	if !e.Prepared() || e.Tasks[1].Setup != Ready {
		t.Errorf("got %+v, want both ready and the clock started", e)
	}
}

func TestResumeKeepsWhatIsAlreadySetUp(t *testing.T) {
	saved := &Exam{Minutes: 30, Tasks: []Task{{ID: "tr-svc", Setup: Ready}, {ID: "wl-scale", Setup: Preparing}}}
	s, _, r := openSession(t, withReset("tr-svc", "wl-scale"), saved)

	s.Resume(func() {})
	waitStarted(t, s)

	if got := r.ran("reset.sh"); !slices.Equal(got, []string{"wl-scale"}) {
		t.Errorf("reset ran for %v, want only wl-scale: tr-svc is set up and must stay broken", got)
	}
}

func TestResumeFinishesTheLastTasks(t *testing.T) {
	started := now.Add(-10 * time.Minute)
	saved := &Exam{Minutes: 30, Started: started, Deadline: started.Add(30 * time.Minute),
		Tasks: []Task{{ID: "wl-scale", Setup: Ready}, {ID: "tr-svc", Setup: Preparing}}}
	s, _, r := newSession(t, saved)

	if !s.Resume(func() {}) {
		t.Fatal("Resume had nothing to do")
	}
	e := waitSetUp(t, s)

	if got := r.ran("setup.sh"); !slices.Equal(got, []string{"tr-svc"}) {
		t.Errorf("setup ran for %v, want only tr-svc", got)
	}
	if !e.Started.Equal(started) {
		t.Errorf("started %v, want the clock kept at %v", e.Started, started)
	}
}

func TestResumeLeavesAStartedExamAlone(t *testing.T) {
	saved := &Exam{Minutes: 30, Started: now, Deadline: now.Add(30 * time.Minute), Tasks: []Task{{ID: "wl-scale", Setup: Ready}}}
	s, _, r := newSession(t, saved)

	if s.Resume(func() { t.Error("released for nothing") }) {
		t.Error("Resume had work to do on a started exam")
	}
	if len(r.calls) != 0 {
		t.Errorf("ran %v", r.calls)
	}
}

func TestOpenRefusesAnExamWithAnUnknownTask(t *testing.T) {
	store := &memStore{exam: &Exam{Tasks: []Task{{ID: "gone"}}}}

	_, err := Open(Config{Store: store, Catalog: catalog(t, files)})

	if err == nil || !strings.Contains(err.Error(), `"gone"`) {
		t.Errorf("got %v", err)
	}
}

func TestDiscard(t *testing.T) {
	s, store, _ := newSession(t, nil)
	begin(t, s, 1)

	if err := s.Discard(); err != nil {
		t.Fatal(err)
	}

	if _, ok := s.Current(); ok || store.exam != nil {
		t.Error("the exam is still there")
	}
}

var now = time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

var files = fstest.MapFS{
	"wl-scale/task.md":     {Data: []byte("---\nid: wl-scale\ntitle: Scale\nhost: node-1\ndomain: workloads\nweight: 4\n---\nScale it.\n")},
	"tr-svc/task.md":       {Data: []byte("---\nid: tr-svc\ntitle: Service\nhost: node-2\ndomain: troubleshooting\nweight: 6\n---\nFix it.\n")},
	"wl-scale/setup.sh":    {Data: []byte("setup.sh\n")},
	"wl-scale/check.sh":    {Data: []byte("check.sh\n")},
	"wl-scale/solution.sh": {Data: []byte("solution.sh\n")},
	"wl-scale/explain.md":  {Data: []byte("Why.\n")},
	"tr-svc/setup.sh":      {Data: []byte("setup.sh\n")},
	"tr-svc/check.sh":      {Data: []byte("check.sh\n")},
	"tr-svc/solution.sh":   {Data: []byte("solution.sh\n")},
	"tr-svc/explain.md":    {Data: []byte("Why.\n")},
}

func withReset(ids ...string) fstest.MapFS {
	fsys := maps.Clone(files)
	for _, id := range ids {
		fsys[id+"/reset.sh"] = &fstest.MapFile{Data: []byte("reset.sh\n")}
	}
	return fsys
}

// withLast adds ar-etcd, a task set up after the rest.
func withLast() fstest.MapFS {
	fsys := maps.Clone(files)
	fsys["ar-etcd/task.md"] = &fstest.MapFile{Data: []byte("---\nid: ar-etcd\ntitle: Restore\nhost: node-1\ndomain: architecture\nweight: 8\norder: last\n---\nRestore it.\n")}
	for _, name := range []string{"setup.sh", "check.sh", "solution.sh", "explain.md"} {
		fsys["ar-etcd/"+name] = files["wl-scale/"+name]
	}
	return fsys
}

func catalog(t *testing.T, fsys fstest.MapFS) []tasks.Task {
	all, err := tasks.Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	return all
}

func begin(t *testing.T, s *Session, n int) Exam {
	if err := s.Begin(s.cfg.Catalog, n, 30, func() {}); err != nil {
		t.Fatal(err)
	}
	return waitStarted(t, s)
}

func end(t *testing.T, s *Session) {
	if _, err := s.End(func() {}); err != nil {
		t.Fatal(err)
	}
}

func waitScored(t *testing.T, s *Session) Exam {
	for range 500 {
		if e, _ := s.Current(); e.Over() && !e.Scoring() {
			return e
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the exam was never scored")
	return Exam{}
}

func waitStarted(t *testing.T, s *Session) Exam {
	for range 500 {
		if e, _ := s.Current(); e.Prepared() {
			return e
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the clock never started")
	return Exam{}
}

func waitSetUp(t *testing.T, s *Session) Exam {
	for range 500 {
		if e, _ := s.Current(); e.Prepared() && !e.SettingUp() {
			return e
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the setups never finished")
	return Exam{}
}

func newSession(t *testing.T, saved *Exam) (*Session, *memStore, *fakeRunner) {
	return openSession(t, files, saved)
}

func openSession(t *testing.T, fsys fstest.MapFS, saved *Exam) (*Session, *memStore, *fakeRunner) {
	store := &memStore{exam: saved}
	r := &fakeRunner{}
	s, err := Open(Config{
		Store:   store,
		Files:   fsys,
		Runner:  r,
		Catalog: catalog(t, fsys),
		Now:     func() time.Time { return now },
		Rand:    rand.New(rand.NewPCG(1, 2)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return s, store, r
}

type memStore struct {
	mu   sync.Mutex
	exam *Exam
}

func (m *memStore) Load() (*Exam, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exam, nil
}

func (m *memStore) Save(e Exam) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e.Tasks = slices.Clone(e.Tasks)
	m.exam = &e
	return nil
}

func (m *memStore) saved() Exam {
	m.mu.Lock()
	defer m.mu.Unlock()
	return *m.exam
}

func (m *memStore) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exam = nil
	return nil
}

type call struct{ taskID, script string }

// fakeRunner answers by script name: each fake script's body is its own name, and tasks.TidyScript
// is "tidy". Setups other than ar-etcd's wait until hold is closed, ar-etcd's until lastHold is,
// and checks until checkHold is.
type fakeRunner struct {
	mu        sync.Mutex
	hold      chan struct{}
	lastHold  chan struct{}
	checkHold chan struct{}
	calls     []call
	setupErr  map[string]error
	checkOut  map[string]string
	checkErr  error
}

func (r *fakeRunner) Run(_ context.Context, _, taskID string, script []byte) (string, error) {
	name := strings.TrimSpace(string(script))
	if string(script) == tasks.TidyScript {
		name = "tidy"
	}
	r.mu.Lock()
	r.calls = append(r.calls, call{taskID, name})
	hold, lastHold, checkHold := r.hold, r.lastHold, r.checkHold
	r.mu.Unlock()
	if hold != nil && name == "setup.sh" && taskID != "ar-etcd" {
		<-hold
	}
	if lastHold != nil && name == "setup.sh" && taskID == "ar-etcd" {
		<-lastHold
	}
	if checkHold != nil && name == "check.sh" {
		<-checkHold
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	switch name {
	case "setup.sh":
		return "", r.setupErr[taskID]
	case "check.sh":
		return r.checkOut[taskID], r.checkErr
	}
	return "", nil
}

func (r *fakeRunner) holdChecks() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkHold = make(chan struct{})
}

func (r *fakeRunner) releaseChecks() { close(r.checkHold) }

func (r *fakeRunner) holdLast() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastHold = make(chan struct{})
}

func (r *fakeRunner) releaseLast() { close(r.lastHold) }

func (r *fakeRunner) order() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var got []string
	for _, c := range r.calls {
		got = append(got, c.taskID+"/"+c.script)
	}
	return got
}

func (r *fakeRunner) forget() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = nil
}

func (r *fakeRunner) ran(script string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var ids []string
	for _, c := range r.calls {
		if c.script == script {
			ids = append(ids, c.taskID)
		}
	}
	slices.Sort(ids)
	return ids
}
