package e2e

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"google.golang.org/protobuf/types/known/timestamppb"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestALabStartsWhenAsked(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_ABSENT})

	if got := labStatus(t, http.MethodGet, http.StatusOK); got.State != "none" {
		t.Fatalf("before starting: %+v", got)
	}
	started := labStatus(t, http.MethodPost, http.StatusAccepted)
	if started.State != "starting" || started.Step != "Preparing nodes" || started.StartedAt.IsZero() {
		t.Fatalf("after POST: %+v", started)
	}
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY, Started: timestamppb.New(started.StartedAt), Ready: timestamppb.Now()})

	if got := labStatus(t, http.MethodGet, http.StatusOK); got.State != "ready" || got.ReadyAt.IsZero() {
		t.Errorf("once built: %+v", got)
	}
}

func TestALabStops(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})

	if got := labStatus(t, http.MethodDelete, http.StatusAccepted); got.State != "stopping" {
		t.Errorf("after DELETE: %+v", got)
	}
}

func TestStoppingAStartingLabIsAConflict(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})

	status, body := request(t, http.MethodDelete, "/api/lab", nil)

	if status != http.StatusConflict || !strings.Contains(string(body), `"code":"lab_busy"`) {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestAFailedLabSaysSo(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_FAILED, Error: "the cluster couldn't be created; the lab's log has the cause"})

	if got := labStatus(t, http.MethodGet, http.StatusOK); got.State != "failed" || got.Error == "" {
		t.Errorf("got %+v", got)
	}
}

func TestALabIsUnavailableWhenNoAgentAnswers(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := "http://" + ln.Addr().String()
	ln.Close()
	server, url, err := start(closed)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stop(server) })

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		status, body := requestTo(t, url, method, "/api/lab", nil)

		if status != http.StatusServiceUnavailable || !strings.Contains(string(body), `"code":"lab_unavailable"`) ||
			strings.Contains(string(body), ln.Addr().String()) {
			t.Errorf("%s: status %d: %s", method, status, body)
		}
	}
}

func TestACrossSiteLabStartIsRefused(t *testing.T) {
	status, _ := request(t, http.MethodPost, "/api/lab", map[string]string{"Sec-Fetch-Site": "cross-site"})

	if status != http.StatusForbidden {
		t.Errorf("status %d, want 403", status)
	}
}

type lab struct {
	State     string    `json:"state"`
	Step      string    `json:"step"`
	StartedAt time.Time `json:"startedAt"`
	ReadyAt   time.Time `json:"readyAt"`
	Error     string    `json:"error"`
}

func labStatus(t *testing.T, method string, want int) lab {
	t.Helper()
	status, body := request(t, method, "/api/lab", nil)
	if status != want {
		t.Fatalf("%s /api/lab: status %d: %s", method, status, body)
	}
	var got lab
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

type fakeAgent struct {
	labv1connect.UnimplementedLabServiceHandler
	mu      sync.Mutex
	cluster *labv1.Cluster
}

func newFakeAgent() *fakeAgent {
	return &fakeAgent{cluster: &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_ABSENT}}
}

func (a *fakeAgent) set(c *labv1.Cluster) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cluster = c
}

func (a *fakeAgent) GetCluster(context.Context, *labv1.GetClusterRequest) (*labv1.GetClusterResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return &labv1.GetClusterResponse{Cluster: a.cluster}, nil
}

func (a *fakeAgent) CreateCluster(context.Context, *labv1.CreateClusterRequest) (*labv1.CreateClusterResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch a.cluster.State {
	case labv1.ClusterState_CLUSTER_STATE_DELETING:
		return nil, connect.NewError(connect.CodeFailedPrecondition, "the cluster is still being deleted")
	case labv1.ClusterState_CLUSTER_STATE_ABSENT, labv1.ClusterState_CLUSTER_STATE_FAILED:
		a.cluster = &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING, Step: "Preparing nodes", Started: timestamppb.Now()}
	}
	return &labv1.CreateClusterResponse{Cluster: a.cluster}, nil
}

func (a *fakeAgent) DeleteCluster(context.Context, *labv1.DeleteClusterRequest) (*labv1.DeleteClusterResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch a.cluster.State {
	case labv1.ClusterState_CLUSTER_STATE_CREATING:
		return nil, connect.NewError(connect.CodeFailedPrecondition, "the cluster is still being created")
	case labv1.ClusterState_CLUSTER_STATE_READY, labv1.ClusterState_CLUSTER_STATE_FAILED:
		a.cluster = &labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_DELETING}
	}
	return &labv1.DeleteClusterResponse{Cluster: a.cluster}, nil
}

func (a *fakeAgent) OpenShell(_ context.Context, stream labv1connect.LabServiceOpenShellServerStream) error {
	a.mu.Lock()
	state := a.cluster.State
	a.mu.Unlock()
	if state != labv1.ClusterState_CLUSTER_STATE_READY {
		return connect.NewError(connect.CodeFailedPrecondition, "the cluster isn't ready")
	}
	if _, err := stream.Receive(); err != nil {
		return err
	}
	if err := stream.Send(&labv1.OpenShellResponse{Output: []byte("root@cka-sim-control-plane:~# ")}); err != nil {
		return err
	}
	for {
		req, err := stream.Receive()
		if err != nil {
			return nil
		}
		if keys := req.GetKeys(); keys != nil {
			if err := stream.Send(&labv1.OpenShellResponse{Output: keys}); err != nil {
				return err
			}
		}
	}
}

func (a *fakeAgent) RunScript(_ context.Context, req *labv1.RunScriptRequest) (*labv1.RunScriptResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cluster.State != labv1.ClusterState_CLUSTER_STATE_READY {
		return nil, connect.NewError(connect.CodeFailedPrecondition, "the cluster isn't ready")
	}
	if !strings.Contains(req.GetScript(), "export TASK_ID=") || !strings.Contains(req.GetScript(), "fresh_ns()") {
		return &labv1.RunScriptResponse{Output: []byte("not a catalog script\n"), ExitCode: 2}, nil
	}
	return &labv1.RunScriptResponse{Output: []byte("ran on " + req.GetHost() + "\n")}, nil
}

func serveAgent(a *fakeAgent) (string, func()) {
	server := connect.NewServer(rpcauth.RequireBearer(agentToken))
	labv1connect.RegisterLabServiceHandler(server, a)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.Protocols = new(http.Protocols)
	srv.Config.Protocols.SetHTTP1(true)
	srv.Config.Protocols.SetUnencryptedHTTP2(true)
	srv.Start()
	return srv.URL, srv.Close
}
