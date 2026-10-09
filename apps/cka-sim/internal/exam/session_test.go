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

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

func TestBeginSetsEveryTaskUpThenStartsTheClock(t *testing.T) {
	s, store, r := newSession(t, nil)

	released := make(chan struct{})
	if err := s.Begin(2, 30, func() { close(released) }); err != nil {
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
	fsys := maps.Clone(files)
	fsys["ar-etcd/task.md"] = &fstest.MapFile{Data: []byte("---\nid: ar-etcd\ntitle: Restore\nhost: node-1\ndomain: architecture\nweight: 8\norder: last\n---\nRestore it.\n")}
	for _, name := range []string{"setup.sh", "check.sh", "solution.sh", "explain.md"} {
		fsys["ar-etcd/"+name] = files["wl-scale/"+name]
	}
	s, _, r := openSession(t, fsys, nil)
	r.hold = make(chan struct{})

	if err := s.Begin(3, 30, func() {}); err != nil {
		t.Fatal(err)
	}
	for len(r.ran("setup.sh")) < 2 {
		time.Sleep(time.Millisecond)
	}
	if got := r.ran("setup.sh"); slices.Contains(got, "ar-etcd") {
		t.Errorf("ar-etcd set up alongside %v, want it after them", got)
	}
	close(r.hold)
	waitStarted(t, s)

	if got := r.ran("setup.sh"); len(got) != 3 || !slices.Contains(got, "ar-etcd") {
		t.Errorf("setup ran for %v, want all three", got)
	}
}

func TestBeginRefusesASecondExam(t *testing.T) {
	s, _, _ := newSession(t, nil)
	begin(t, s, 1)

	if err := s.Begin(1, 30, func() {}); !errors.Is(err, ErrExists) {
		t.Errorf("got %v, want ErrExists", err)
	}
}

func TestEndGradesEveryTask(t *testing.T) {
	s, store, r := newSession(t, nil)
	r.checkOut = map[string]string{"wl-scale": "PASS 2 scaled\nPASS 2 ready\n", "tr-svc": "FAIL 3 endpoints\n"}
	begin(t, s, 2)

	e, err := s.End(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if !e.Over() || !store.exam.Over() {
		t.Error("the exam is not over, or not saved as over")
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

func TestEndRecordsACrashedCheck(t *testing.T) {
	s, _, r := newSession(t, nil)
	begin(t, s, 1)
	r.checkErr = errors.New("kubectl: not found")

	e, err := s.End(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if task := e.Tasks[0]; task.Result == nil || task.Result.Total != 0 || !strings.Contains(task.CheckError, "kubectl: not found") {
		t.Errorf("got %+v, want 0 points and the check's error", task)
	}
}

func TestEndRefuses(t *testing.T) {
	t.Run("no exam", func(t *testing.T) {
		s, _, _ := newSession(t, nil)
		if _, err := s.End(context.Background()); !errors.Is(err, ErrNoExam) {
			t.Errorf("got %v", err)
		}
	})
	t.Run("still preparing", func(t *testing.T) {
		s, _, _ := newSession(t, &Exam{Minutes: 30, Tasks: []Task{{ID: "wl-scale", Setup: Preparing}}})
		if _, err := s.End(context.Background()); err == nil || !strings.Contains(err.Error(), "still being set up") {
			t.Errorf("got %v", err)
		}
	})
	t.Run("twice", func(t *testing.T) {
		s, _, _ := newSession(t, nil)
		begin(t, s, 1)
		s.End(context.Background())
		if _, err := s.End(context.Background()); err == nil || !strings.Contains(err.Error(), "already ended") {
			t.Errorf("got %v", err)
		}
	})
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

func catalog(t *testing.T, fsys fstest.MapFS) []tasks.Task {
	all, err := tasks.Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	return all
}

func begin(t *testing.T, s *Session, n int) Exam {
	if err := s.Begin(n, 30, func() {}); err != nil {
		t.Fatal(err)
	}
	return waitStarted(t, s)
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

func (m *memStore) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exam = nil
	return nil
}

type call struct{ taskID, script string }

// fakeRunner answers by script name: each fake script's body is its own name. Setups other than
// ar-etcd's wait until hold is closed.
type fakeRunner struct {
	mu       sync.Mutex
	hold     chan struct{}
	calls    []call
	setupErr map[string]error
	checkOut map[string]string
	checkErr error
}

func (r *fakeRunner) Run(_ context.Context, _, taskID string, script []byte) (string, error) {
	name := strings.TrimSpace(string(script))
	r.mu.Lock()
	r.calls = append(r.calls, call{taskID, name})
	hold := r.hold
	r.mu.Unlock()
	if hold != nil && name == "setup.sh" && taskID != "ar-etcd" {
		<-hold
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
