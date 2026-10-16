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

	listed := listTasks(t)

	if len(listed) != len(files) || len(listed) == 0 {
		t.Errorf("listed %d tasks, the catalog has %d task.md files", len(listed), len(files))
	}
	if !slices.IsSortedFunc(listed, func(a, b listedTask) int { return strings.Compare(a.ID, b.ID) }) {
		t.Error("the tasks are not sorted by id")
	}
}

func TestATaskIsListedWithWhatThePageShows(t *testing.T) {
	listed := listTasks(t)

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

type listedTask struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Domain      string   `json:"domain"`
	DomainTitle string   `json:"domainTitle"`
	Topics      []string `json:"topics"`
	Weight      int      `json:"weight"`
	Host        string   `json:"host"`
}

func listTasks(t *testing.T) []listedTask {
	t.Helper()
	status, body := request(t, http.MethodGet, "/api/tasks", nil)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	var listed []listedTask
	if err := json.Unmarshal(body, &listed); err != nil {
		t.Fatal(err)
	}
	return listed
}
