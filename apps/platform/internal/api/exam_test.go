package api

import (
	"context"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

func TestNoExamIsNotFound(t *testing.T) {
	status, body := call(t, newExamAPI(&fakeRunner{}, nil), http.MethodGet, "/api/exam")

	if status != http.StatusNotFound || !strings.Contains(body, `"code":"no_exam"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestBeginningAnExamAnswersAtOnceWithItsTasksPreparing(t *testing.T) {
	h := newExamAPI(&fakeRunner{}, nil)

	status, body := send(t, h, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)

	if status != http.StatusAccepted || !strings.Contains(body, `"id":"wl-scale"`) || !strings.Contains(body, `"minutes":30`) ||
		!strings.Contains(body, `"now":`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestABadExamRequestIsABadRequest(t *testing.T) {
	for _, tc := range []struct{ body, code string }{
		{`{"count": 0, "minutes": 30}`, "invalid_exam"},
		{`{"count": 4, "minutes": 241}`, "invalid_exam"},
		{`not json`, "malformed_request"},
	} {
		status, body := send(t, newExamAPI(&fakeRunner{}, nil), http.MethodPost, "/api/exam", tc.body)

		if status != http.StatusBadRequest || !strings.Contains(body, `"code":"`+tc.code+`"`) {
			t.Errorf("%s: status %d: %s", tc.body, status, body)
		}
	}
}

func TestAnExamNeedsAReadyLab(t *testing.T) {
	h := newExamAPI(&fakeRunner{}, &fakeLabProvider{lab: labs.Lab{State: labs.Starting}})

	status, body := send(t, h, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)

	if status != http.StatusConflict || !strings.Contains(body, `"code":"lab_not_ready"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestPracticeWaitsWhileAnExamRuns(t *testing.T) {
	h := newExamAPI(&fakeRunner{}, nil)
	send(t, h, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)

	for _, path := range []string{"/api/tasks/wl-scale/start", "/api/tasks/wl-scale/check", "/api/tasks/wl-scale/tidy"} {
		status, body := call(t, h, http.MethodPost, path)
		if status != http.StatusConflict || !strings.Contains(body, `"code":"exam_running"`) {
			t.Errorf("%s: status %d: %s", path, status, body)
		}
	}

	status, body := call(t, h, http.MethodGet, "/api/tasks/wl-scale/solution")
	if status != http.StatusConflict || strings.Contains(body, "kubectl scale") {
		t.Errorf("solution: status %d: %s", status, body)
	}
}

func TestAnExamIsFlaggedEndedGradedAndDiscarded(t *testing.T) {
	h := newExamAPI(&fakeRunner{result: labs.ScriptResult{Output: "PASS 3 scaled\nFAIL 1 labelled\n"}}, nil)
	send(t, h, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)
	waitFor(t, h, `"deadline":`)

	if status, body := call(t, h, http.MethodPut, "/api/exam/tasks/wl-scale/flag"); status != http.StatusNoContent {
		t.Fatalf("flag: status %d: %s", status, body)
	}
	if status, body := call(t, h, http.MethodPut, "/api/exam/tasks/nw-dns/flag"); status != http.StatusNotFound ||
		!strings.Contains(body, `"code":"task_not_in_exam"`) {
		t.Errorf("flag an unknown task: status %d: %s", status, body)
	}
	if status, body := call(t, h, http.MethodDelete, "/api/exam"); status != http.StatusConflict {
		t.Errorf("discard a running exam: status %d: %s", status, body)
	}

	if status, body := call(t, h, http.MethodPost, "/api/exam/end"); status != http.StatusAccepted || !strings.Contains(body, `"flagged":true`) {
		t.Fatalf("end: status %d: %s", status, body)
	}
	body := waitFor(t, h, `"score":`)

	if !strings.Contains(body, `"score":{"percent":75,"passed":true,"pass":66}`) || !strings.Contains(body, `"earned":3,"total":4`) {
		t.Errorf("graded: %s", body)
	}
	if status, body := call(t, h, http.MethodPost, "/api/exam/end"); status != http.StatusConflict || !strings.Contains(body, `"code":"exam_ended"`) {
		t.Errorf("end twice: status %d: %s", status, body)
	}
	if status, body := call(t, h, http.MethodDelete, "/api/exam"); status != http.StatusNoContent {
		t.Errorf("discard: status %d: %s", status, body)
	}
	if status, _ := call(t, h, http.MethodGet, "/api/exam"); status != http.StatusNotFound {
		t.Errorf("after discard: status %d", status)
	}
}

func TestEndingNoExamIsNotFound(t *testing.T) {
	status, body := call(t, newExamAPI(&fakeRunner{}, nil), http.MethodPost, "/api/exam/end")

	if status != http.StatusNotFound || !strings.Contains(body, `"code":"no_exam"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func newExamAPI(r *fakeRunner, lab *fakeLabProvider) http.Handler {
	if lab == nil {
		lab = &fakeLabProvider{lab: labs.Lab{State: labs.Ready}}
	}
	taskCases := usecases.New(fakeRepository{tasks: []tasks.Task{wlScale}}, r)
	labCases := labusecases.New(lab)
	return New(taskCases, labCases, examsFor(taskCases, labCases, nil), nil)
}

func examsFor(taskCases usecases.UseCases, labCases labusecases.UseCases, repo examusecases.ExamRepository) examusecases.UseCases {
	if repo == nil {
		repo = &memoryExams{}
	}
	return examusecases.New(repo, taskCases, labCases.Get, time.Now, rand.New(rand.NewPCG(1, 2)))
}

func send(t *testing.T, h http.Handler, method, path, body string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body)))
	got, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	return rec.Code, string(got)
}

func waitFor(t *testing.T, h http.Handler, want string) string {
	t.Helper()
	for range 200 {
		_, body := send(t, h, http.MethodGet, "/api/exam", "")
		if strings.Contains(body, want) {
			return body
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("GET /api/exam never had %s", want)
	return ""
}

type memoryExams struct {
	mu   sync.Mutex
	exam *exams.Exam
}

func (m *memoryExams) Load(string) (*exams.Exam, error) { return nil, nil }

func (m *memoryExams) Owners() ([]string, error) { return nil, nil }

func (m *memoryExams) Save(_ string, e exams.Exam) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e = e.Clone()
	m.exam = &e
	return nil
}

func (m *memoryExams) Clear(string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exam = nil
	return nil
}
