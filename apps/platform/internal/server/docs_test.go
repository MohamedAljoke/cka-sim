package server

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/docs"
)

func TestDocsSearchAnswersTheAssistantWithPassagesAndLinks(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := asAssistant(t, srv, http.MethodPost, "/api/docs/search", `{"query": "back up etcd"}`)

	var reply docsReply
	if err := json.Unmarshal([]byte(body), &reply); status != http.StatusOK || err != nil {
		t.Fatalf("got %d %q", status, body)
	}
	if len(reply.Results) != 1 || reply.Results[0].URL != etcdDocs || reply.Results[0].Text != "about back up etcd" {
		t.Errorf("results = %+v", reply.Results)
	}
}

func TestDocsSearchPassesTheTasksDomainOn(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	_, body := asAssistant(t, srv, http.MethodPost, "/api/docs/search", `{"query": "back up etcd", "domain": "architecture"}`)

	var reply docsReply
	if err := json.Unmarshal([]byte(body), &reply); err != nil || len(reply.Results) != 1 ||
		reply.Results[0].Text != "about back up etcd in architecture" {
		t.Errorf("got %q", body)
	}
}

func TestDocsSearchRefusesAMalformedBody(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := request(t, srv, http.MethodPost, "/api/docs/search", `{"query": 3}`)

	if status != http.StatusBadRequest {
		t.Errorf("got %d %q", status, body)
	}
}

func TestDocsSearchPassesOnTheSearchsRefusal(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := request(t, srv, http.MethodPost, "/api/docs/search", `{"query": "nothing indexed"}`)

	if status != http.StatusServiceUnavailable || problemCode(t, body) != "docs_empty" {
		t.Errorf("got %d %q", status, body)
	}
}

func TestDocsSearchSaysWhenItIsOff(t *testing.T) {
	srv := serve(t, readyDesks(t, &fakeRunner{}, newFakeOpener()), Practice{})

	status, body := request(t, srv, http.MethodPost, "/api/docs/search", `{"query": "back up etcd"}`)

	if status != http.StatusServiceUnavailable || problemCode(t, body) != "docs_off" {
		t.Errorf("got %d %q", status, body)
	}
}

const etcdDocs = "https://kubernetes.io/docs/tasks/administer-cluster/configure-upgrade-etcd/"

type fakeDocs struct{}

func (fakeDocs) Find(_ context.Context, query, domain string) ([]docs.Result, error) {
	if query == "nothing indexed" {
		return nil, docs.ErrEmpty
	}
	text := "about " + query
	if domain != "" {
		text += " in " + domain
	}
	return []docs.Result{{Title: "Operating etcd clusters", URL: etcdDocs, Text: text}}, nil
}
