package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/assistant"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/desk"
	labagent "github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
)

type assistantRequest struct {
	Provider string `json:"provider"`
}

func (a *api) startAssistant(w http.ResponseWriter, r *http.Request) error {
	var req assistantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.Malformed(`the body must be JSON like {"provider": "claude"}`)
	}
	if _, _, err := a.prepareAssistant(r, req.Provider); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Prepares again even after the POST: a socket opened on its own must not skip the settings.
func (a *api) serveAssistant(w http.ResponseWriter, r *http.Request) error {
	d, p, err := a.prepareAssistant(r, r.URL.Query().Get("provider"))
	if err != nil {
		return err
	}
	labagent.Serve(w, r, func(ctx context.Context, size labagent.Size) (labagent.Shell, error) {
		return d.OpenAssistant(ctx, p.Command(), size)
	})
	return nil
}

func (a *api) prepareAssistant(r *http.Request, id string) (*desk.Desk, assistant.Provider, error) {
	p, err := assistant.Find(a.Assistants, id)
	if err != nil {
		return nil, nil, err
	}
	d, err := a.desk(r)
	if err != nil {
		return nil, nil, err
	}
	if err := notDuringExam(d); err != nil {
		return nil, nil, err
	}
	if err := p.Prepare(r.Context(), d.Lab); err != nil {
		return nil, nil, scriptFailed(err)
	}
	return d, p, nil
}
