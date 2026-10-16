package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type labView struct {
	State     labs.State `json:"state"`
	Step      string     `json:"step,omitempty"`
	StartedAt time.Time  `json:"startedAt,omitzero"`
	ReadyAt   time.Time  `json:"readyAt,omitzero"`
	Error     string     `json:"error,omitempty"`
}

func (a *api) getLabStatus(w http.ResponseWriter, r *http.Request) error {
	lab, err := a.getLab.Execute(r.Context())
	switch {
	case errors.Is(err, labs.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, newLabView(lab))
}

func (a *api) postLab(w http.ResponseWriter, r *http.Request) error {
	lab, err := a.startLab.Execute(r.Context())
	switch {
	case errors.Is(err, labs.ErrBusy):
		return httpx.Conflict(err)
	case errors.Is(err, labs.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}
	return httpx.WriteJSON(w, http.StatusAccepted, newLabView(lab))
}

func (a *api) deleteLab(w http.ResponseWriter, r *http.Request) error {
	lab, err := a.stopLab.Execute(r.Context())
	switch {
	case errors.Is(err, labs.ErrBusy):
		return httpx.Conflict(err)
	case errors.Is(err, labs.ErrUnavailable):
		return httpx.Unavailable(err)
	case err != nil:
		return err
	}
	return httpx.WriteJSON(w, http.StatusAccepted, newLabView(lab))
}

func newLabView(lab labs.Lab) labView {
	return labView{State: lab.State, Step: lab.Step, StartedAt: lab.StartedAt, ReadyAt: lab.ReadyAt, Error: lab.Error}
}
