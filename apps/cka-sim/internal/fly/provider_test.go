package fly

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
)

func TestCreateResumesAWarmMachine(t *testing.T) {
	api := newFakeAPI(t, machine{ID: "m1", PrivateIP: "fdaa::2", Pool: "ready"})
	p := newProvider(api, nil)

	box, err := p.Create(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if box.ID != "m1" {
		t.Errorf("box = %q, want m1", box.ID)
	}
	if got := p.hosts(); !slices.Equal(got, []string{"tcp://[fdaa::2]:2375"}) {
		t.Errorf("docker hosts = %q, want the machine's private address", got)
	}
	want := []string{"POST /machines/m1/metadata/pool claimed", "POST /machines/m1/start", "GET /machines/m1/wait"}
	if got := api.writes(); !slices.Equal(got, want) {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestCreateReplacesAnUnhealthyMachine(t *testing.T) {
	api := newFakeAPI(t,
		machine{ID: "cold", PrivateIP: "fdaa::2", Pool: "ready"},
		machine{ID: "warm", PrivateIP: "fdaa::3", Pool: "ready"})
	p := newProvider(api, map[string]error{"cold": errors.New("no cluster")})

	box, err := p.Create(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if box.ID != "warm" {
		t.Errorf("box = %q, want warm", box.ID)
	}
	if !slices.Contains(api.writes(), "DELETE /machines/cold") {
		t.Errorf("calls = %q, want the unhealthy machine destroyed", api.writes())
	}
}

func TestCreateWithAnEmptyPoolFails(t *testing.T) {
	api := newFakeAPI(t, machine{ID: "busy", PrivateIP: "fdaa::2", Pool: "claimed"})

	_, err := newProvider(api, nil).Create(context.Background())

	if !errors.Is(err, ErrPoolEmpty) {
		t.Fatalf("err = %v, want ErrPoolEmpty", err)
	}
}

func TestDestroyDeletesTheMachine(t *testing.T) {
	api := newFakeAPI(t, machine{ID: "m1", PrivateIP: "fdaa::2", Pool: "claimed"})

	if err := newProvider(api, nil).Destroy(context.Background(), sandbox.Box{ID: "m1"}); err != nil {
		t.Fatal(err)
	}

	if got := api.writes(); !slices.Equal(got, []string{"DELETE /machines/m1"}) {
		t.Errorf("calls = %q, want one delete", got)
	}
}

func TestReapDestroysClaimedMachines(t *testing.T) {
	api := newFakeAPI(t,
		machine{ID: "left", PrivateIP: "fdaa::2", Pool: "claimed"},
		machine{ID: "warm", PrivateIP: "fdaa::3", Pool: "ready"})

	if err := newProvider(api, nil).ReapClaimed(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := api.writes(); !slices.Equal(got, []string{"DELETE /machines/left"}) {
		t.Errorf("calls = %q, want only the claimed machine destroyed", got)
	}
}

// testProvider records the docker hosts it was asked to reach, and fails the probe for unhealthy ids.
type testProvider struct {
	*Provider
	mu   sync.Mutex
	seen []string
}

func newProvider(api *fakeAPI, unhealthy map[string]error) *testProvider {
	tp := &testProvider{}
	tp.Provider = New(Config{App: "labs", Token: "tok", API: api.srv.URL})
	tp.healthWait = 50 * time.Millisecond
	tp.box = func(ctx context.Context, id, host string) (sandbox.Box, error) {
		tp.mu.Lock()
		defer tp.mu.Unlock()
		tp.seen = append(tp.seen, host)
		return sandbox.Box{ID: id}, nil
	}
	tp.probe = func(ctx context.Context, b sandbox.Box) error { return unhealthy[b.ID] }
	return tp
}

func (tp *testProvider) hosts() []string {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return slices.Clone(tp.seen)
}

type machine struct {
	ID        string
	PrivateIP string
	Pool      string
}

// fakeAPI is the slice of the Machines API the provider uses. It records every call except lists.
type fakeAPI struct {
	srv      *httptest.Server
	mu       sync.Mutex
	machines []machine
	calls    []string
}

func newFakeAPI(t *testing.T, ms ...machine) *fakeAPI {
	api := &fakeAPI{machines: ms}
	api.srv = httptest.NewServer(http.HandlerFunc(api.serve))
	t.Cleanup(api.srv.Close)
	return api
}

func (api *fakeAPI) serve(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer tok" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	path, ok := strings.CutPrefix(r.URL.Path, "/v1/apps/labs")
	if !ok {
		http.NotFound(w, r)
		return
	}
	api.mu.Lock()
	defer api.mu.Unlock()
	if r.Method == http.MethodGet && path == "/machines" {
		api.list(w, r.URL.Query().Get("metadata.pool"))
		return
	}
	call := r.Method + " " + path
	if strings.HasSuffix(path, "/metadata/pool") {
		var body struct{ Value string }
		json.NewDecoder(r.Body).Decode(&body)
		call += " " + body.Value
		api.setPool(strings.Split(path, "/")[2], body.Value)
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.Write([]byte(`{"ok":true}`))
	}
	api.calls = append(api.calls, call)
}

func (api *fakeAPI) list(w http.ResponseWriter, pool string) {
	type listed struct {
		ID        string `json:"id"`
		PrivateIP string `json:"private_ip"`
	}
	var out []listed
	for _, m := range api.machines {
		if m.Pool == pool {
			out = append(out, listed{ID: m.ID, PrivateIP: m.PrivateIP})
		}
	}
	json.NewEncoder(w).Encode(out)
}

func (api *fakeAPI) setPool(id, pool string) {
	for i := range api.machines {
		if api.machines[i].ID == id {
			api.machines[i].Pool = pool
		}
	}
}

func (api *fakeAPI) writes() []string {
	api.mu.Lock()
	defer api.mu.Unlock()
	return slices.Clone(api.calls)
}
