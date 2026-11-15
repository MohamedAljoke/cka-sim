package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	studyv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1/studyv1connect"
	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

var nwDNS = tasks.Task{ID: "nw-dns", Title: "Resolve a Service", Domain: tasks.Networking, Topics: []string{"dns", "services"}, Weight: 3, Host: "node-1", Question: "Find its IP.", Setup: "fresh_ns nw-dns\n", Check: "check 1 \"found\" true\n"}

func TestListTasksAnswersTheMatchingTasksWithoutTheirQuestions(t *testing.T) {
	res, err := studyClient(t, newStudyAPI(&fakeRunner{}, &fakeAttempts{}, nil), tutorToken).ListTasks(context.Background(),
		&studyv1.ListTasksRequest{Student: "local", Domains: []string{"networking"}, Topics: []string{"dns"}})

	if err != nil || len(res.GetTasks()) != 1 {
		t.Fatalf("got %v, %v", res, err)
	}
	got := res.GetTasks()[0]
	if got.GetId() != "nw-dns" || got.GetTitle() != "Resolve a Service" || got.GetWeight() != 3 ||
		!slices.Equal(got.GetTopics(), []string{"dns", "services"}) || strings.Contains(res.String(), "Find its IP") {
		t.Errorf("got %v", res)
	}
}

func TestListTasksRefusesAnUnknownDomainNamingIt(t *testing.T) {
	_, err := studyClient(t, newStudyAPI(&fakeRunner{}, &fakeAttempts{}, nil), tutorToken).ListTasks(context.Background(),
		&studyv1.ListTasksRequest{Student: "local", Domains: []string{"security"}})

	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), `"security"`) {
		t.Errorf("got %v", err)
	}
}

func TestStartTaskTidiesTheTaskThePageSetUpThenSetsUpTheNewOne(t *testing.T) {
	r := &fakeRunner{}
	h := newStudyAPI(r, &fakeAttempts{}, nil)
	call(t, h, http.MethodPost, "/api/tasks/wl-scale/start")

	res, err := studyClient(t, h, tutorToken).StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "nw-dns"})

	if err != nil || !res.GetReady() || res.GetTask().GetQuestion() != "Find its IP." || res.GetTask().GetHost() != "node-1" {
		t.Fatalf("got %v, %v", res, err)
	}
	if !slices.Equal(r.ran, []string{"fresh_ns wl-scale\n", usecases.TidyScript("wl-scale"), "fresh_ns nw-dns\n"}) {
		t.Errorf("ran %q", r.ran)
	}
	current, err := tutorClient(t, h, tutorToken).GetCurrentTask(context.Background(), &tutorv1.GetCurrentTaskRequest{Student: "local"})
	if err != nil || current.GetTask().GetId() != "nw-dns" {
		t.Errorf("current %v, %v", current, err)
	}
}

func TestAFailedSetupIsAnAnswerWithWhy(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{ExitCode: 1, Output: "no such image\n"}}

	res, err := studyClient(t, newStudyAPI(r, &fakeAttempts{}, nil), tutorToken).
		StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "nw-dns"})

	if err != nil || res.GetReady() || res.GetExitCode() != 1 || res.GetOutput() != "no such image\n" {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestStudyStartingAnUnknownTaskIsNotFound(t *testing.T) {
	_, err := studyClient(t, newStudyAPI(&fakeRunner{}, &fakeAttempts{}, nil), tutorToken).
		StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "no-such-task"})

	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "task_not_found") {
		t.Errorf("got %v", err)
	}
}

func TestStartingATaskWaitsForTheLab(t *testing.T) {
	_, err := studyClient(t, newStudyAPI(&fakeRunner{err: labs.ErrNotReady}, &fakeAttempts{}, nil), tutorToken).
		StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "nw-dns"})

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "lab_not_ready") {
		t.Errorf("got %v", err)
	}
}

func TestEveryPracticeCheckIsInTheHistoryNewestFirst(t *testing.T) {
	r := &fakeRunner{result: labs.ScriptResult{Output: "PASS 1 scaled\nFAIL 1 labelled\n"}}
	h := newStudyAPI(r, &fakeAttempts{}, nil)
	call(t, h, http.MethodPost, "/api/tasks/wl-scale/check")
	r.result = labs.ScriptResult{Output: "PASS 1 found\n"}
	tutorClient(t, h, tutorToken).CheckTask(context.Background(), &tutorv1.CheckTaskRequest{Student: "local", TaskId: "nw-dns"})

	res, err := studyClient(t, h, tutorToken).GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local"})

	if err != nil || len(res.GetEntries()) != 2 {
		t.Fatalf("got %v, %v", res, err)
	}
	newest, oldest := res.GetEntries()[0].GetAttempt(), res.GetEntries()[1].GetAttempt()
	if newest.GetTaskId() != "nw-dns" || newest.GetEarned() != 1 || len(newest.GetFailed()) != 0 {
		t.Errorf("newest %v", newest)
	}
	if oldest.GetTaskId() != "wl-scale" || oldest.GetTotal() != 2 || !slices.Equal(oldest.GetFailed(), []string{"labelled"}) ||
		oldest.GetAt().AsTime().IsZero() {
		t.Errorf("oldest %v", oldest)
	}
}

func TestTheHistoryOfSomeTasksOnly(t *testing.T) {
	attempts := &fakeAttempts{entries: []tasks.Entry{{Kind: tasks.AttemptEntry, TaskID: "wl-scale"}, {Kind: tasks.AttemptEntry, TaskID: "nw-dns"}}}

	res, err := studyClient(t, newStudyAPI(&fakeRunner{}, attempts, nil), tutorToken).
		GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local", TaskIds: []string{"wl-scale"}})

	if err != nil || len(res.GetEntries()) != 1 || res.GetEntries()[0].GetAttempt().GetTaskId() != "wl-scale" {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestANoteSavedIsInTheHistory(t *testing.T) {
	client := studyClient(t, newStudyAPI(&fakeRunner{}, &fakeAttempts{}, nil), tutorToken)

	_, err := client.SaveNote(context.Background(), &studyv1.SaveNoteRequest{Student: "local", TaskId: "nw-dns", Text: "mixed up from and to"})
	res, _ := client.GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local"})

	if err != nil || len(res.GetEntries()) != 1 || res.GetEntries()[0].GetNote().GetText() != "mixed up from and to" {
		t.Errorf("got %v, %v", res, err)
	}
}

func TestANoteOver500CharactersIsAnInvalidArgument(t *testing.T) {
	attempts := &fakeAttempts{}

	_, err := studyClient(t, newStudyAPI(&fakeRunner{}, attempts, nil), tutorToken).SaveNote(context.Background(),
		&studyv1.SaveNoteRequest{Student: "local", TaskId: "nw-dns", Text: strings.Repeat("x", 501)})

	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "note_too_long") || len(attempts.entries) != 0 {
		t.Errorf("got %v; saved %v", err, attempts.entries)
	}
}

func TestANoteOnAnUnknownTaskIsNotFound(t *testing.T) {
	_, err := studyClient(t, newStudyAPI(&fakeRunner{}, &fakeAttempts{}, nil), tutorToken).SaveNote(context.Background(),
		&studyv1.SaveNoteRequest{Student: "local", TaskId: "no-such-task", Text: "a note"})

	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "task_not_found") {
		t.Errorf("got %v", err)
	}
}

func TestAHistoryThatCannotBeReadIsUnavailableAndHidesTheCause(t *testing.T) {
	attempts := &fakeAttempts{err: errors.New("read /var/cka/history: input/output error")}

	_, err := studyClient(t, newStudyAPI(&fakeRunner{}, attempts, nil), tutorToken).
		GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local"})

	if connect.CodeOf(err) != connect.CodeUnavailable || !strings.Contains(err.Error(), "history_unavailable") ||
		strings.Contains(err.Error(), "/var/cka") {
		t.Errorf("got %v", err)
	}
}

func TestStudyRefusesAnUnknownStudent(t *testing.T) {
	h := newStudyAPI(&fakeRunner{}, &fakeAttempts{}, map[string]string{"token-of-alice": "alice"})

	_, err := studyClient(t, h, tutorToken).GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local"})

	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "unknown_student") {
		t.Errorf("got %v", err)
	}
}

func TestStudyWaitsWhileAnExamRuns(t *testing.T) {
	attempts := &fakeAttempts{}
	h := newStudyAPI(&fakeRunner{}, attempts, nil)
	send(t, h, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)
	client := studyClient(t, h, tutorToken)

	for name, call := range map[string]func() error{
		"ListTasks": func() error {
			_, err := client.ListTasks(context.Background(), &studyv1.ListTasksRequest{Student: "local"})
			return err
		},
		"StartTask": func() error {
			_, err := client.StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "nw-dns"})
			return err
		},
		"GetHistory": func() error {
			_, err := client.GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local"})
			return err
		},
		"SaveNote": func() error {
			_, err := client.SaveNote(context.Background(), &studyv1.SaveNoteRequest{Student: "local", TaskId: "nw-dns", Text: "a"})
			return err
		},
	} {
		if err := call(); connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "exam_running") {
			t.Errorf("%s: %v", name, err)
		}
	}
	if len(attempts.entries) != 0 {
		t.Errorf("saved %v", attempts.entries)
	}
}

func TestStudyNeedsTheToken(t *testing.T) {
	_, err := studyClient(t, newStudyAPI(&fakeRunner{}, &fakeAttempts{}, nil), "not-the-token").
		ListTasks(context.Background(), &studyv1.ListTasksRequest{Student: "local"})

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func newStudyAPI(r *fakeRunner, attempts *fakeAttempts, students map[string]string) http.Handler {
	taskCases := usecases.New(fakeRepository{tasks: []tasks.Task{wlScale, nwDNS}}, r, attempts, time.Now)
	labCases := labusecases.New(&fakeLabProvider{lab: labs.Lab{State: labs.Ready}}, r)
	examCases := examsFor(taskCases, labCases, nil)
	page := New(taskCases, labCases, examCases, chatusecases.UseCases{}, students)

	return WithTutor(page, tutorToken, taskCases, labCases, examCases, students)
}

func studyClient(t *testing.T, h http.Handler, token string) studyv1connect.StudyServiceClient {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return studyv1connect.NewStudyServiceClient(connect.NewClient(connecthttp.NewTransport(srv.Client(), srv.URL), rpcauth.Bearer(token)))
}

type fakeAttempts struct {
	mu      sync.Mutex
	entries []tasks.Entry
	err     error
}

func (a *fakeAttempts) Append(_ context.Context, _ string, entry tasks.Entry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.err != nil {
		return a.err
	}
	a.entries = append(a.entries, entry)
	return nil
}

func (a *fakeAttempts) List(context.Context, string) ([]tasks.Entry, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return slices.Clone(a.entries), a.err
}
