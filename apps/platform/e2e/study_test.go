package e2e

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	studyv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1/studyv1connect"
	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1/tutorv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestStudyListsTheCatalogWithoutTheQuestions(t *testing.T) {
	res, err := studyClient(baseURL, assistantToken).ListTasks(context.Background(),
		&studyv1.ListTasksRequest{Student: "local", Topics: []string{"poddisruptionbudgets"}})

	ids := []string{}
	for _, task := range res.GetTasks() {
		ids = append(ids, task.GetId())
	}
	if err != nil || !slices.Contains(ids, "wl-pdb") || slices.Contains(ids, "nw-nodeport") ||
		strings.Contains(res.String(), "never drop") {
		t.Errorf("got %v, %v", ids, err)
	}
}

func TestStudyStartTaskTidiesTheTaskThePageSetUp(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())
	requestTo(t, url, http.MethodPost, "/api/tasks/wl-pdb/start", nil)
	ran := agent.scriptsFrom()

	res, err := studyClient(url, assistantToken).StartTask(context.Background(),
		&studyv1.StartTaskRequest{Student: "local", TaskId: "nw-nodeport"})

	if err != nil || !res.GetReady() || res.GetTask().GetHost() != "cka-sim-control-plane" || res.GetTask().GetQuestion() == "" {
		t.Fatalf("got %v, %v", res, err)
	}
	scripts := ran()
	tidied := slices.Index(scripts, "kubectl delete namespace -l cka-sim/task=wl-pdb --wait=false >/dev/null\n")
	setUp := slices.IndexFunc(scripts, func(s string) bool { return strings.Contains(s, "export TASK_ID='nw-nodeport'") })
	if tidied < 0 || setUp < tidied {
		t.Errorf("ran %q", scripts)
	}
	current, err := tutorClientTo(url).GetCurrentTask(context.Background(), &tutorv1.GetCurrentTaskRequest{Student: "local"})
	if err != nil || current.GetTask().GetId() != "nw-nodeport" {
		t.Errorf("current %v, %v", current, err)
	}
}

func TestEveryPracticeCheckIsKeptAndOutlivesARestart(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	dir := t.TempDir()
	server, url, err := start(agentURL, dir)
	if err != nil {
		t.Fatal(err)
	}
	requestTo(t, url, http.MethodPost, "/api/tasks/wl-pdb/check", nil)
	tutorClientTo(url).CheckTask(context.Background(), &tutorv1.CheckTaskRequest{Student: "local", TaskId: "nw-nodeport"})
	if _, err := studyClient(url, assistantToken).SaveNote(context.Background(),
		&studyv1.SaveNoteRequest{Student: "local", TaskId: "nw-nodeport", Text: "forgot the nodePort"}); err != nil {
		t.Fatal(err)
	}
	if err := stop(server); err != nil {
		t.Fatal(err)
	}

	url = ownServer(t, dir)
	res, err := studyClient(url, assistantToken).GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local"})

	if err != nil || len(res.GetEntries()) != 3 {
		t.Fatalf("got %v, %v", res, err)
	}
	note, newer, older := res.GetEntries()[0].GetNote(), res.GetEntries()[1].GetAttempt(), res.GetEntries()[2].GetAttempt()
	if note.GetText() != "forgot the nodePort" || newer.GetTaskId() != "nw-nodeport" || older.GetTaskId() != "wl-pdb" ||
		older.GetEarned() != 2 || older.GetTotal() != 3 || !slices.Equal(older.GetFailed(), []string{"not done yet"}) {
		t.Errorf("got %v", res)
	}
}

func TestStudyWaitsForTheExamAndTheExamLeavesNoAttempts(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())
	study := studyClient(url, assistantToken)
	post(t, url, "/api/exam", `{"count": 2, "minutes": 30}`)
	waitForExam(t, url, func(e exam) bool { return e.Deadline != "" })

	_, listErr := study.ListTasks(context.Background(), &studyv1.ListTasksRequest{Student: "local"})
	_, startErr := study.StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "wl-pdb"})
	_, historyErr := study.GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local"})
	_, noteErr := study.SaveNote(context.Background(), &studyv1.SaveNoteRequest{Student: "local", TaskId: "wl-pdb", Text: "a"})
	for _, err := range []error{listErr, startErr, historyErr, noteErr} {
		if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "exam_running") {
			t.Errorf("during the exam: %v", err)
		}
	}

	requestTo(t, url, http.MethodPost, "/api/exam/end", nil)
	waitForExam(t, url, func(e exam) bool { return e.Score != nil })
	res, err := study.GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "local"})
	if err != nil || len(res.GetEntries()) != 0 {
		t.Errorf("after the exam: %v, %v", res, err)
	}
}

func TestStudyRefusesWhatItCannotDo(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	study := studyClient(baseURL, assistantToken)

	_, err := study.StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "no-such-task"})
	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "task_not_found") {
		t.Errorf("start an unknown task: %v", err)
	}
	_, err = study.SaveNote(context.Background(), &studyv1.SaveNoteRequest{Student: "local", TaskId: "no-such-task", Text: "a"})
	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "task_not_found") {
		t.Errorf("a note on an unknown task: %v", err)
	}
	_, err = study.SaveNote(context.Background(), &studyv1.SaveNoteRequest{Student: "local", TaskId: "wl-pdb", Text: strings.Repeat("x", 501)})
	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "note_too_long") {
		t.Errorf("a long note: %v", err)
	}
	_, err = study.SaveNote(context.Background(), &studyv1.SaveNoteRequest{Student: "local", TaskId: "wl-pdb", Text: "  "})
	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "empty_note") {
		t.Errorf("an empty note: %v", err)
	}
	_, err = study.ListTasks(context.Background(), &studyv1.ListTasksRequest{Student: "local", Domains: []string{"security"}})
	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "unknown_domain") {
		t.Errorf("an unknown domain: %v", err)
	}
	_, err = study.GetHistory(context.Background(), &studyv1.GetHistoryRequest{Student: "mallory"})
	if connect.CodeOf(err) != connect.CodeNotFound || !strings.Contains(err.Error(), "unknown_student") {
		t.Errorf("an unknown student: %v", err)
	}
	_, err = studyClient(baseURL, agentToken).ListTasks(context.Background(), &studyv1.ListTasksRequest{Student: "local"})
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("the agent's token: %v", err)
	}

	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})
	_, err = study.StartTask(context.Background(), &studyv1.StartTaskRequest{Student: "local", TaskId: "wl-pdb"})
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "lab_not_ready") {
		t.Errorf("a lab that isn't ready: %v", err)
	}
}

func studyClient(url, token string) studyv1connect.StudyServiceClient {
	return studyv1connect.NewStudyServiceClient(
		connect.NewClient(connecthttp.NewTransport(http.DefaultClient, url), rpcauth.Bearer(token)))
}

func tutorClientTo(url string) tutorv1connect.TutorServiceClient {
	return tutorv1connect.NewTutorServiceClient(
		connect.NewClient(connecthttp.NewTransport(http.DefaultClient, url), rpcauth.Bearer(assistantToken)))
}
