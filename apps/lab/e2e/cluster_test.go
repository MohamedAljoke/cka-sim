package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestTheAgentIsHealthy(t *testing.T) {
	res, err := http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		t.Errorf("status %d", res.StatusCode)
	}
}

func TestEveryCallIsRefusedWithoutTheRightToken(t *testing.T) {
	for name, c := range map[string]func() error{
		"get without a token": func() error {
			_, err := client().GetCluster(context.Background(), &labv1.GetClusterRequest{})
			return err
		},
		"create with a wrong token": func() error {
			_, err := client(rpcauth.Bearer("guess")).CreateCluster(context.Background(), &labv1.CreateClusterRequest{})
			return err
		},
		"delete with a wrong token": func() error {
			_, err := client(rpcauth.Bearer("guess")).DeleteCluster(context.Background(), &labv1.DeleteClusterRequest{})
			return err
		},
		"run a script with a wrong token": func() error {
			_, err := client(rpcauth.Bearer("guess")).RunScript(context.Background(), &labv1.RunScriptRequest{Host: controlPlane, Script: "true"})
			return err
		},
	} {
		if err := c(); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

func TestAFreshAgentHasNoCluster(t *testing.T) {
	if *create {
		t.Skip("the -create run changes the cluster")
	}

	got, err := authed().GetCluster(context.Background(), &labv1.GetClusterRequest{})

	if err != nil || got.Cluster.State != labv1.ClusterState_CLUSTER_STATE_ABSENT {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestDeletingNoClusterChangesNothing(t *testing.T) {
	if *create {
		t.Skip("the -create run changes the cluster")
	}

	got, err := authed().DeleteCluster(context.Background(), &labv1.DeleteClusterRequest{})

	if err != nil || got.Cluster.State != labv1.ClusterState_CLUSTER_STATE_ABSENT {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestARealClusterIsBuiltAndDeleted(t *testing.T) {
	if !*create {
		t.Skip("needs -create")
	}
	c := authed()

	created, err := c.CreateCluster(context.Background(), &labv1.CreateClusterRequest{})
	if err != nil || created.Cluster.State != labv1.ClusterState_CLUSTER_STATE_CREATING {
		t.Fatalf("CreateCluster = %v, %v", created, err)
	}
	if _, err := c.DeleteCluster(context.Background(), &labv1.DeleteClusterRequest{}); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("delete while creating: got %v", err)
	}
	ready := waitWhile(t, labv1.ClusterState_CLUSTER_STATE_CREATING, 20*time.Minute)
	if ready.State != labv1.ClusterState_CLUSTER_STATE_READY {
		t.Fatalf("the build ended %v", ready)
	}
	t.Logf("ready in %s", ready.Ready.AsTime().Sub(ready.Started.AsTime()).Round(time.Second))

	runInShell(t, "echo user=$(whoami) kubectl=$(command -v kubectl || echo none)\n", "user=candidate kubectl=none")
	output := runInShell(t, "ssh "+worker+" 'echo nodes=$(kubectl get nodes --no-headers | grep -cw Ready)'\n", "nodes=3")
	t.Logf("the shell answered %q", output)
	runScripts(t)

	if _, err := c.DeleteCluster(context.Background(), &labv1.DeleteClusterRequest{}); err != nil {
		t.Fatal(err)
	}
	if gone := waitWhile(t, labv1.ClusterState_CLUSTER_STATE_DELETING, 5*time.Minute); gone.State != labv1.ClusterState_CLUSTER_STATE_ABSENT {
		t.Errorf("the delete ended %v", gone)
	}
}

func waitWhile(t *testing.T, state labv1.ClusterState, limit time.Duration) *labv1.Cluster {
	t.Helper()
	deadline := time.Now().Add(limit)
	step := ""
	for {
		got, err := authed().GetCluster(context.Background(), &labv1.GetClusterRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Cluster.State != state {
			return got.Cluster
		}
		if got.Cluster.Step != step {
			step = got.Cluster.Step
			t.Log(step)
		}
		if time.Now().After(deadline) {
			t.Fatalf("still %v after %s", got.Cluster, limit)
		}
		time.Sleep(time.Second)
	}
}
