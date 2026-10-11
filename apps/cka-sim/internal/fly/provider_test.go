package fly

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
)

func TestCreateResumesAWarmMachine(t *testing.T) {
	api := newFakeAPI(t, machine{ID: "m1", PrivateIP: "fdaa::2", Pool: "ready"})
	p := newProvider(t, api, nil)

	box, err := p.Create(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if box.ID != "m1" {
		t.Errorf("box = %q, want m1", box.ID)
	}
	r := p.relays["m1"]
	if r == nil || r.target != "[fdaa::2]:2375" {
		t.Fatalf("relay = %+v, want one to the machine's private address", r)
	}
	if got := p.hosts(); !slices.Equal(got, []string{r.host()}) {
		t.Errorf("docker hosts = %q, want the relay %q", got, r.host())
	}
	want := []string{
		"POST /machines/m1/metadata/pool claimed",
		"POST /network_policies cka-sim-lock-m1 m1",
		"POST /machines/m1/start",
		"GET /machines/m1/wait",
	}
	if got := api.writes(); !slices.Equal(got, want) {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestCreateReplacesAnUnhealthyMachine(t *testing.T) {
	api := newFakeAPI(t,
		machine{ID: "cold", PrivateIP: "fdaa::2", Pool: "ready"},
		machine{ID: "warm", PrivateIP: "fdaa::3", Pool: "ready"})
	p := newProvider(t, api, map[string]error{"cold": errors.New("no cluster")})

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
	if got := api.policyNames(); !slices.Equal(got, []string{"cka-sim-lock-warm"}) {
		t.Errorf("policies = %q, want only the claimed machine's lock", got)
	}
}

func TestCreateNeverResumesAMachineItCannotLock(t *testing.T) {
	api := newFakeAPI(t,
		machine{ID: "open", PrivateIP: "fdaa::2", Pool: "ready"},
		machine{ID: "warm", PrivateIP: "fdaa::3", Pool: "ready"})
	api.lockFails = "open"
	p := newProvider(t, api, nil)

	box, err := p.Create(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if box.ID != "warm" {
		t.Errorf("box = %q, want warm", box.ID)
	}
	if got := api.writes(); slices.Contains(got, "POST /machines/open/start") || !slices.Contains(got, "DELETE /machines/open") {
		t.Errorf("calls = %q, want the unlockable machine destroyed without being resumed", got)
	}
}

func TestCreateWithAnEmptyPoolFails(t *testing.T) {
	api := newFakeAPI(t, machine{ID: "busy", PrivateIP: "fdaa::2", Pool: "claimed"})

	_, err := newProvider(t, api, nil).Create(context.Background())

	if !errors.Is(err, ErrPoolEmpty) {
		t.Fatalf("err = %v, want ErrPoolEmpty", err)
	}
}

func TestDestroyDeletesTheMachineThenItsLock(t *testing.T) {
	api := newFakeAPI(t,
		machine{ID: "m1", PrivateIP: "fdaa::2", Pool: "claimed"},
		machine{ID: "m2", PrivateIP: "fdaa::3", Pool: "claimed"})
	api.policies = []policy{{ID: "p1", Name: "cka-sim-lock-m1"}, {ID: "p2", Name: "cka-sim-lock-m2"}}

	if err := newProvider(t, api, nil).Destroy(context.Background(), sandbox.Box{ID: "m1"}); err != nil {
		t.Fatal(err)
	}

	if got, want := api.writes(), []string{"DELETE /machines/m1", "DELETE /network_policies/p1"}; !slices.Equal(got, want) {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestDestroyClosesTheRelay(t *testing.T) {
	api := newFakeAPI(t, machine{ID: "m1", PrivateIP: "fdaa::2", Pool: "ready"})
	p := newProvider(t, api, nil)
	box, err := p.Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	path := p.relays["m1"].path

	if err := p.Destroy(context.Background(), box); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("relay socket %s still there: %v", path, err)
	}
	if len(p.relays) != 0 {
		t.Errorf("relays = %v, want none", p.relays)
	}
}

func TestReapDestroysClaimedMachinesAndEveryLock(t *testing.T) {
	api := newFakeAPI(t,
		machine{ID: "left", PrivateIP: "fdaa::2", Pool: "claimed"},
		machine{ID: "warm", PrivateIP: "fdaa::3", Pool: "ready"})
	api.policies = []policy{
		{ID: "p1", Name: "cka-sim-lock-left"},
		{ID: "p2", Name: "cka-sim-lock-gone"},
		{ID: "p3", Name: "someone-elses"},
	}

	if err := newProvider(t, api, nil).ReapClaimed(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := []string{"DELETE /machines/left", "DELETE /network_policies/p1", "DELETE /network_policies/p2"}
	if got := api.writes(); !slices.Equal(got, want) {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

// testProvider records the docker hosts it was asked to reach, and fails the probe for unhealthy ids.
type testProvider struct {
	*Provider
	mu   sync.Mutex
	seen []string
}

func newProvider(t *testing.T, api *fakeAPI, unhealthy map[string]error) *testProvider {
	t.Setenv("TMPDIR", t.TempDir())
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
	t.Cleanup(func() {
		for _, r := range tp.relays {
			r.Close()
		}
	})
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
	policies []policy
	// lockFails is a machine whose network policy can't be written.
	lockFails string
	calls     []string
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
	if r.Method == http.MethodGet && path == "/network_policies" {
		json.NewEncoder(w).Encode(api.policies)
		return
	}
	call := r.Method + " " + path
	if r.Method == http.MethodPost && path == "/network_policies" {
		name, id := decodeLock(r)
		api.calls = append(api.calls, call+" "+name+" "+id)
		if id == api.lockFails {
			http.Error(w, "policy rejected", http.StatusInternalServerError)
			return
		}
		api.policies = append(api.policies, policy{ID: "pol-" + id, Name: name})
		w.Write([]byte(`{"id":"pol-` + id + `"}`))
		return
	}
	if r.Method == http.MethodDelete && strings.HasPrefix(path, "/network_policies/") {
		id := strings.TrimPrefix(path, "/network_policies/")
		api.policies = slices.DeleteFunc(api.policies, func(p policy) bool { return p.ID == id })
	}
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

func (api *fakeAPI) policyNames() []string {
	api.mu.Lock()
	defer api.mu.Unlock()
	var names []string
	for _, p := range api.policies {
		names = append(names, p.Name)
	}
	return names
}

// decodeLock returns a lock's name and the one machine it selects, and checks it denies all egress.
func decodeLock(r *http.Request) (name, machineID string) {
	var body struct {
		Name     string
		Selector struct{ Machines []struct{ ID string } }
		Rules    []struct {
			Action, Direction string
			Ports             []struct{ Port int }
		}
	}
	json.NewDecoder(r.Body).Decode(&body)
	denyAll := len(body.Rules) == 1 && body.Rules[0].Direction == "egress" &&
		len(body.Rules[0].Ports) == 1 && body.Rules[0].Ports[0].Port == 0
	if len(body.Selector.Machines) != 1 || !denyAll {
		return body.Name, "malformed"
	}
	return body.Name, body.Selector.Machines[0].ID
}
