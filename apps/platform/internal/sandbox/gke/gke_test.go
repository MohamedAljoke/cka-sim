package gke

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

const claims = "/apis/extensions.agents.x-k8s.io/v1beta1/namespaces/labs/sandboxclaims"

func TestCreateClaimsALabAndReachesItsPod(t *testing.T) {
	api := newFakeAPI(t)
	api.readyAfter = 2
	p := newProvider(api, nil)

	box, err := p.Create(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	if box.ID != "lab-1" {
		t.Errorf("box = %q, want the claim lab-1", box.ID)
	}
	if got := p.ips(); !slices.Equal(got, []string{"10.80.0.5"}) {
		t.Errorf("agent addresses = %q, want the sandbox pod's IP", got)
	}
	created := api.created()
	if created.Spec.WarmPoolRef.Name != "labs" || created.Metadata.Labels[managedBy] != ours {
		t.Errorf("claim = %+v, want pool labs, labelled as ours", created)
	}
}

func TestCreateDeletesAClaimThatNeverAnswers(t *testing.T) {
	api := newFakeAPI(t)
	p := newProvider(api, errors.New("connection refused"))

	_, err := p.Create(context.Background())

	if err == nil || !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("err = %v, want the probe's failure", err)
	}
	if !slices.Contains(api.calls(), "DELETE "+claims+"/lab-1") {
		t.Errorf("calls = %q, want the claim deleted", api.calls())
	}
}

func TestCreateGivesUpWhenTheClaimNeverGetsReady(t *testing.T) {
	api := newFakeAPI(t)
	api.readyAfter = 1000
	p := newProvider(api, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := p.Create(ctx)

	if err == nil || !strings.Contains(err.Error(), "never became ready") {
		t.Fatalf("err = %v, want never became ready", err)
	}
	if !slices.Contains(api.calls(), "DELETE "+claims+"/lab-1") {
		t.Errorf("calls = %q, want the claim deleted", api.calls())
	}
}

func TestDestroyDeletesTheClaim(t *testing.T) {
	api := newFakeAPI(t)

	if err := newProvider(api, nil).Destroy(context.Background(), sandbox.Box{ID: "lab-7"}); err != nil {
		t.Fatal(err)
	}

	if got := api.calls(); !slices.Equal(got, []string{"DELETE " + claims + "/lab-7"}) {
		t.Errorf("calls = %q, want one delete", got)
	}
}

func TestReapDeletesOurLeftoverClaims(t *testing.T) {
	api := newFakeAPI(t)
	api.leftover = []string{"lab-a", "lab-b"}

	if err := newProvider(api, nil).ReapClaims(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"GET " + claims + "?labelSelector=app.kubernetes.io%2Fmanaged-by%3Dcka-sim",
		"DELETE " + claims + "/lab-a",
		"DELETE " + claims + "/lab-b",
	}
	if got := api.calls(); !slices.Equal(got, want) {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestRequestsCarryTheCurrentToken(t *testing.T) {
	api := newFakeAPI(t)
	p := newProvider(api, nil)
	p.kube.token = func() (string, error) { return "rotated", nil }

	p.Destroy(context.Background(), sandbox.Box{ID: "lab-1"})

	if api.lastAuth != "Bearer rotated" {
		t.Errorf("Authorization = %q, want the token read for this call", api.lastAuth)
	}
}

type createdClaim struct {
	Metadata struct {
		GenerateName string            `json:"generateName"`
		Labels       map[string]string `json:"labels"`
	} `json:"metadata"`
	Spec struct {
		WarmPoolRef struct {
			Name string `json:"name"`
		} `json:"warmPoolRef"`
	} `json:"spec"`
}

type fakeAPI struct {
	mu         sync.Mutex
	log        []string
	body       createdClaim
	polls      int
	readyAfter int
	leftover   []string
	lastAuth   string
	url        string
}

func newFakeAPI(t *testing.T) *fakeAPI {
	api := &fakeAPI{}
	srv := httptest.NewServer(http.HandlerFunc(api.serve))
	t.Cleanup(srv.Close)
	api.url = srv.URL
	return api
}

func (a *fakeAPI) serve(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastAuth = r.Header.Get("Authorization")
	call := r.Method + " " + r.URL.RequestURI()
	switch {
	case r.Method == http.MethodPost && r.URL.Path == claims:
		json.NewDecoder(r.Body).Decode(&a.body)
		a.log = append(a.log, "POST "+claims)
		fmt.Fprint(w, `{"metadata":{"name":"lab-1"}}`)
	case r.Method == http.MethodGet && r.URL.Path == claims+"/lab-1":
		a.polls++
		if a.polls < a.readyAfter {
			fmt.Fprint(w, `{"status":{"conditions":[{"type":"Ready","status":"False"}]}}`)
			return
		}
		fmt.Fprint(w, `{"status":{"conditions":[{"type":"Ready","status":"True"}],"sandbox":{"name":"labs-x"}}}`)
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/namespaces/labs/pods/labs-x":
		fmt.Fprint(w, `{"status":{"podIP":"10.80.0.5"}}`)
	case r.Method == http.MethodGet && r.URL.Path == claims:
		a.log = append(a.log, call)
		items := make([]string, len(a.leftover))
		for i, name := range a.leftover {
			items[i] = `{"metadata":{"name":"` + name + `"}}`
		}
		fmt.Fprint(w, `{"items":[`+strings.Join(items, ",")+`]}`)
	case r.Method == http.MethodDelete:
		a.log = append(a.log, call)
		fmt.Fprint(w, `{}`)
	default:
		http.Error(w, "unexpected "+call, http.StatusNotFound)
	}
}

func (a *fakeAPI) calls() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return slices.Clone(a.log)
}

func (a *fakeAPI) created() createdClaim {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.body
}

type testProvider struct {
	*Provider
	mu      sync.Mutex
	reached []string
}

func (p *testProvider) ips() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.reached)
}

func newProvider(api *fakeAPI, probeErr error) *testProvider {
	tp := &testProvider{}
	tp.Provider = &Provider{
		kube: kube{
			base:      api.url,
			namespace: "labs",
			token:     func() (string, error) { return "t", nil },
			client:    http.DefaultClient,
		},
		pool:       "labs",
		poll:       time.Millisecond,
		healthWait: 20 * time.Millisecond,
		box: func(id, ip string) sandbox.Box {
			tp.mu.Lock()
			defer tp.mu.Unlock()
			tp.reached = append(tp.reached, ip)
			return sandbox.Box{ID: id}
		},
		probe: func(context.Context, sandbox.Box) error { return probeErr },
	}
	return tp
}
