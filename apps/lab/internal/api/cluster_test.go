package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster/usecases"
)

const token = "s3cret"

func TestHealthNeedsNoToken(t *testing.T) {
	url := serve(t, &fakeProvisioner{})

	res, err := http.Get(url + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		t.Errorf("status %d", res.StatusCode)
	}
}

func TestEveryCallNeedsTheToken(t *testing.T) {
	c := labv1connect.NewLabServiceClient(connect.NewClient(connecthttp.NewTransport(http.DefaultClient, serve(t, &fakeProvisioner{}))))

	_, err := c.GetCluster(context.Background(), &labv1.GetClusterRequest{})

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func TestANewLabHasNoCluster(t *testing.T) {
	got, err := client(serve(t, &fakeProvisioner{})).GetCluster(context.Background(), &labv1.GetClusterRequest{})

	if err != nil || got.Cluster.State != labv1.ClusterState_CLUSTER_STATE_ABSENT || got.Cluster.Started != nil {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestCreateClusterAnswersCreatingAndThenReady(t *testing.T) {
	p := &fakeProvisioner{created: make(chan error)}
	c := client(serve(t, p))

	got, err := c.CreateCluster(context.Background(), &labv1.CreateClusterRequest{})
	if err != nil || got.Cluster.State != labv1.ClusterState_CLUSTER_STATE_CREATING || got.Cluster.Started == nil {
		t.Fatalf("CreateCluster = %v, %v", got, err)
	}
	p.created <- nil

	ready := waitForState(t, c, labv1.ClusterState_CLUSTER_STATE_READY)
	if ready.Ready == nil || ready.Ready.AsTime().Before(ready.Started.AsTime()) {
		t.Errorf("ready cluster %v", ready)
	}
}

func TestDeleteClusterWhileCreatingFailsItsPrecondition(t *testing.T) {
	p := &fakeProvisioner{created: make(chan error)}
	c := client(serve(t, p))
	c.CreateCluster(context.Background(), &labv1.CreateClusterRequest{})
	t.Cleanup(func() { p.created <- nil })

	_, err := c.DeleteCluster(context.Background(), &labv1.DeleteClusterRequest{})

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeFailedPrecondition || connectErr.Message() != cluster.ErrBusy.Message {
		t.Errorf("got %v", err)
	}
}

func TestAFailedBuildShowsAFailedClusterWithoutItsCause(t *testing.T) {
	p := &fakeProvisioner{created: make(chan error)}
	c := client(serve(t, p))
	c.CreateCluster(context.Background(), &labv1.CreateClusterRequest{})
	p.created <- errors.New("pull kindest/node from /secret/registry: denied")

	failed := waitForState(t, c, labv1.ClusterState_CLUSTER_STATE_FAILED)

	if failed.Error == "" || strings.Contains(failed.Error, "secret") {
		t.Errorf("error = %q", failed.Error)
	}
}

func TestDeleteClusterAnswersDeletingAndThenAbsent(t *testing.T) {
	p := &fakeProvisioner{created: make(chan error), deleted: make(chan error)}
	c := client(serve(t, p))
	c.CreateCluster(context.Background(), &labv1.CreateClusterRequest{})
	p.created <- nil
	waitForState(t, c, labv1.ClusterState_CLUSTER_STATE_READY)

	got, err := c.DeleteCluster(context.Background(), &labv1.DeleteClusterRequest{})
	if err != nil || got.Cluster.State != labv1.ClusterState_CLUSTER_STATE_DELETING {
		t.Fatalf("DeleteCluster = %v, %v", got, err)
	}
	p.deleted <- nil

	waitForState(t, c, labv1.ClusterState_CLUSTER_STATE_ABSENT)
}

func TestTheServiceAlsoAnswersPlainJSON(t *testing.T) {
	url := serve(t, &fakeProvisioner{})
	req, err := http.NewRequest(http.MethodPost, url+labv1connect.LabServiceGetClusterProcedure, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `"state":"CLUSTER_STATE_ABSENT"`) {
		t.Errorf("status %d: %s", res.StatusCode, body)
	}
}

type fakeProvisioner struct {
	created chan error
	deleted chan error
}

func (p *fakeProvisioner) Create(context.Context, func(string)) error { return <-p.created }

func (p *fakeProvisioner) Delete(context.Context) error { return <-p.deleted }

func serve(t *testing.T, p usecases.Provisioner) string {
	t.Helper()
	return serveWith(t, p, newEchoOpener())
}

func client(url string) labv1connect.LabServiceClient {
	return labv1connect.NewLabServiceClient(connect.NewClient(connecthttp.NewTransport(http.DefaultClient, url), rpcauth.Bearer(token)))
}

func waitForState(t *testing.T, c labv1connect.LabServiceClient, want labv1.ClusterState) *labv1.Cluster {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		got, err := c.GetCluster(context.Background(), &labv1.GetClusterRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Cluster.State == want {
			return got.Cluster
		}
		if time.Now().After(deadline) {
			t.Fatalf("still %v after 5s, want %v", got.Cluster, want)
		}
		time.Sleep(time.Millisecond)
	}
}
