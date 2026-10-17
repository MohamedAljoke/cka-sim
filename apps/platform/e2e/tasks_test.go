package e2e

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestTheTaskListHasEveryTaskInTheCatalog(t *testing.T) {
	files, err := filepath.Glob("../catalog/*/task.md")
	if err != nil {
		t.Fatal(err)
	}

	listed := listTasks(t, "")

	if len(listed) != len(files) || len(listed) == 0 {
		t.Errorf("listed %d tasks, the catalog has %d task.md files", len(listed), len(files))
	}
	byTitle := func(a, b listedTask) int { return strings.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title)) }
	if !slices.IsSortedFunc(listed, byTitle) {
		t.Error("the tasks are not sorted by title")
	}
}

func TestATaskIsListedWithWhatThePageShows(t *testing.T) {
	listed := listTasks(t, "")

	i := slices.IndexFunc(listed, func(task listedTask) bool { return task.ID == "wl-scale" })
	if i < 0 {
		t.Fatal("wl-scale is not listed")
	}
	got := listed[i]
	if got.Title != "Scale a Deployment" || got.Domain != "workloads" || got.DomainTitle != "Workloads and Scheduling" ||
		!slices.Equal(got.Topics, []string{"deployments", "scaling"}) || got.Weight != 4 || got.Host != "cka-sim-control-plane" {
		t.Errorf("wl-scale = %+v", got)
	}
}

func TestTheTaskListNeverCarriesTheQuestion(t *testing.T) {
	status, body := request(t, http.MethodGet, "/api/tasks", nil)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}

	var raw []map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}

	for _, task := range raw {
		if _, ok := task["question"]; ok {
			t.Fatalf("a listed task carries its question: %s", task["id"])
		}
		if string(task["topics"]) == "null" {
			t.Errorf("task %s has null topics, want an array", task["id"])
		}
	}
}

func TestTheListFiltersByDomainAndTopic(t *testing.T) {
	listed := listTasks(t, "?domain=networking&topic=networkpolicy")

	if len(listed) == 0 {
		t.Fatal("no networking task has the topic networkpolicy")
	}
	for _, task := range listed {
		if task.Domain != "networking" || !slices.Contains(task.Topics, "networkpolicy") {
			t.Errorf("%s does not match: %+v", task.ID, task)
		}
	}
}

func TestRepeatedDomainsKeepTasksInAnyOfThem(t *testing.T) {
	listed := listTasks(t, "?domain=networking&domain=storage")

	seen := map[string]bool{}
	for _, task := range listed {
		seen[task.Domain] = true
	}
	if len(seen) != 2 || !seen["networking"] || !seen["storage"] {
		t.Errorf("domains listed: %v", seen)
	}
}

func TestTheListSearchesIDsAndTitles(t *testing.T) {
	listed := listTasks(t, "?q=NGINX%20pod")

	if len(listed) != 1 || listed[0].ID != "wl-pod" {
		t.Errorf("listed %+v, want only wl-pod", listed)
	}
}

func TestASearchNoTaskMatchesListsNothing(t *testing.T) {
	if listed := listTasks(t, "?q=no-such-task"); len(listed) != 0 {
		t.Errorf("listed %d tasks", len(listed))
	}
}

func TestATopicNoTaskHasListsNothing(t *testing.T) {
	if listed := listTasks(t, "?topic=no-such-topic"); len(listed) != 0 {
		t.Errorf("listed %d tasks", len(listed))
	}
}

func TestABadFilterIsRefused(t *testing.T) {
	for _, tc := range []struct{ query, code string }{
		{"?domain=security", "unknown_domain"},
		{"?domain=Networking", "unknown_domain"},
		{"?topic=Network%20Policy", "malformed_topic"},
	} {
		status, body := request(t, http.MethodGet, "/api/tasks"+tc.query, nil)

		if status != http.StatusBadRequest || !strings.Contains(string(body), `"code":"`+tc.code+`"`) {
			t.Errorf("%s: status %d: %s", tc.query, status, body)
		}
	}
}

func TestAnUnknownRouteIsNotFound(t *testing.T) {
	status, _ := request(t, http.MethodGet, "/api/nope", nil)

	if status != http.StatusNotFound {
		t.Errorf("status %d, want 404", status)
	}
}

func TestTheTaskListIsReadOnly(t *testing.T) {
	status, _ := request(t, http.MethodPost, "/api/tasks", nil)

	if status != http.StatusMethodNotAllowed {
		t.Errorf("status %d, want 405", status)
	}
}

func TestACrossSiteWriteIsRefused(t *testing.T) {
	status, _ := request(t, http.MethodPost, "/api/tasks", map[string]string{"Sec-Fetch-Site": "cross-site"})

	if status != http.StatusForbidden {
		t.Errorf("status %d, want 403", status)
	}
}

func TestATaskOpensWithItsQuestion(t *testing.T) {
	status, body := request(t, http.MethodGet, "/api/tasks/wl-scale", nil)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}

	var got taskDetail
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}

	if got.ID != "wl-scale" || got.Title != "Scale a Deployment" || got.DomainTitle != "Workloads and Scheduling" ||
		!strings.HasPrefix(got.Question, "## Task") || !strings.Contains(got.Question, "Scale it to **4** replicas") {
		t.Errorf("wl-scale = %+v", got)
	}
}

func TestEveryListedTaskOpensWithAQuestion(t *testing.T) {
	for _, task := range listTasks(t, "") {
		status, body := request(t, http.MethodGet, "/api/tasks/"+task.ID, nil)
		var got taskDetail
		if err := json.Unmarshal(body, &got); status != http.StatusOK || err != nil || strings.TrimSpace(got.Question) == "" {
			t.Errorf("%s: status %d, err %v, question %q", task.ID, status, err, got.Question)
		}
	}
}

func TestAnUnknownTaskIsNotFound(t *testing.T) {
	for _, path := range []string{"/api/tasks/no-such-task", "/api/tasks/..%2Fcatalog", "/api/tasks/WL-SCALE"} {
		status, body := request(t, http.MethodGet, path, nil)

		if status != http.StatusNotFound || !strings.Contains(string(body), `"code":"task_not_found"`) {
			t.Errorf("%s: status %d: %s", path, status, body)
		}
	}
}

func TestATaskOpensWithoutItsSolution(t *testing.T) {
	status, body := request(t, http.MethodGet, "/api/tasks/wl-scale", nil)

	if status != http.StatusOK || strings.Contains(string(body), "solution") || strings.Contains(string(body), "kubectl -n wl-scale scale") {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestATaskShowsItsSolutionWhenAsked(t *testing.T) {
	status, body := request(t, http.MethodGet, "/api/tasks/wl-scale/solution", nil)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}

	var got solution
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got.Solution, "kubectl -n wl-scale scale deployment web --replicas=4") {
		t.Errorf("solution = %q", got.Solution)
	}
}

func TestEveryListedTaskHasASolution(t *testing.T) {
	for _, task := range listTasks(t, "") {
		status, body := request(t, http.MethodGet, "/api/tasks/"+task.ID+"/solution", nil)
		var got solution
		if err := json.Unmarshal(body, &got); status != http.StatusOK || err != nil || strings.TrimSpace(got.Solution) == "" {
			t.Errorf("%s: status %d, err %v, solution %q", task.ID, status, err, got.Solution)
		}
	}
}

func TestTheSolutionOfAnUnknownTaskIsNotFound(t *testing.T) {
	status, body := request(t, http.MethodGet, "/api/tasks/no-such-task/solution", nil)

	if status != http.StatusNotFound || !strings.Contains(string(body), `"code":"task_not_found"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

type listedTask struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Domain      string   `json:"domain"`
	DomainTitle string   `json:"domainTitle"`
	Topics      []string `json:"topics"`
	Weight      int      `json:"weight"`
	Host        string   `json:"host"`
}

type taskDetail struct {
	listedTask
	Question string `json:"question"`
}

type solution struct {
	Solution string `json:"solution"`
}

func listTasks(t *testing.T, query string) []listedTask {
	t.Helper()
	status, body := request(t, http.MethodGet, "/api/tasks"+query, nil)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	var listed []listedTask
	if err := json.Unmarshal(body, &listed); err != nil {
		t.Fatal(err)
	}
	return listed
}
