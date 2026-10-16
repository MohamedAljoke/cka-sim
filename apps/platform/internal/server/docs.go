package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/docs"
)

const docsBody = 4 << 10

var (
	errDocsOff = domainerr.Unavailable("docs_off", "searching the Kubernetes docs isn't set up on this server")
	errBadDocs = httpx.Malformed(`the body must be JSON like {"query": "how do I back up etcd"}`)
)

type DocsSearch interface {
	Find(ctx context.Context, query, domain string) ([]docs.Result, error)
}

type docsRequest struct {
	Query  string `json:"query"`
	Domain string `json:"domain"`
}

type docsReply struct {
	Results []docs.Result `json:"results"`
}

func (a *api) searchDocs(w http.ResponseWriter, r *http.Request) error {
	if a.Docs == nil {
		return errDocsOff
	}
	var req docsRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, docsBody)).Decode(&req); err != nil {
		return errBadDocs
	}
	results, err := a.Docs.Find(r.Context(), req.Query, req.Domain)
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, docsReply{Results: results})
}
