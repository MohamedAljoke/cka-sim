package e2e

import (
	"context"
	"net/http"
	"os/exec"
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
	checkAddons(t)

	if _, err := c.DeleteCluster(context.Background(), &labv1.DeleteClusterRequest{}); err != nil {
		t.Fatal(err)
	}
	if gone := waitWhile(t, labv1.ClusterState_CLUSTER_STATE_DELETING, 5*time.Minute); gone.State != labv1.ClusterState_CLUSTER_STATE_ABSENT {
		t.Errorf("the delete ended %v", gone)
	}
	for _, name := range []string{clusterName + "-registry", clusterName + "-charts"} {
		if exec.Command("docker", "inspect", name).Run() == nil {
			t.Errorf("%s is still there", name)
		}
	}
}

func checkAddons(t *testing.T) {
	t.Helper()
	for name, script := range map[string]string{
		"helm and its chart repo": "helm repo add lab http://cka-sim-charts >/dev/null && helm search repo lab/web --versions | grep -q 1.1.0",
		"metrics-server":          "for i in $(seq 60); do kubectl top nodes >/dev/null 2>&1 && exit 0; sleep 5; done; exit 1",
		"Gateway API and Traefik": "kubectl get gatewayclass traefik && kubectl -n traefik rollout status deploy/traefik --timeout=120s",
		"VPA": "kubectl get crd verticalpodautoscalers.autoscaling.k8s.io && " +
			"kubectl -n kube-system rollout status deploy/vpa-recommender --timeout=120s",
		"an nginx Ingress through Traefik": `set -e
kubectl create deployment e2e-web --image=nginx:1.27-alpine
kubectl expose deployment e2e-web --port=80
kubectl create ingress e2e-web --class=nginx --rule="e2e.local/*=e2e-web:80"
kubectl rollout status deploy/e2e-web --timeout=120s
for i in $(seq 30); do curl -sf -H 'Host: e2e.local' http://10.96.80.80/ | grep -q nginx && exit 0; sleep 2; done
echo "Traefik never answered for the Ingress"; exit 1`,
		"kube-router and Services without endpoints": `set -e
kubectl create namespace e2e-policy
kubectl -n e2e-policy run web --image=nginx:1.27-alpine --labels=app=web --expose --port=80
kubectl -n e2e-policy run client --image=busybox:1.36 --command -- sleep 1d
kubectl -n e2e-policy wait --for=condition=Ready pod/web pod/client --timeout=120s
kubectl -n kube-system rollout status ds/kube-router --timeout=120s
kubectl -n e2e-policy exec client -- wget -qO- -T 5 http://web >/dev/null
kubectl -n e2e-policy create service clusterip nobody --tcp=80
if ! kubectl -n e2e-policy exec client -- wget -qO- -T 5 http://nobody 2>&1 | grep -q refused; then
  echo "a Service without endpoints wasn't refused"; exit 1
fi
kubectl -n e2e-policy apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata: {name: deny-all}
spec: {podSelector: {}, policyTypes: [Ingress]}
EOF
sleep 5
if kubectl -n e2e-policy exec client -- wget -qO- -T 5 http://web >/dev/null 2>&1; then
  echo "deny-all let traffic through"; exit 1
fi`,
		"web-operator staged": "test -s /etc/cka-sim/addons/web-operator.yaml",
		"private registry": `kubectl create secret docker-registry e2e-pull --docker-server=cka-sim-registry:5000 \
  --docker-username=candidate --docker-password=cka-registry
kubectl run e2e-pull --image=cka-sim-registry:5000/team/web:1.0 --overrides='{"spec":{"imagePullSecrets":[{"name":"e2e-pull"}]}}'
kubectl wait --for=condition=Ready pod/e2e-pull --timeout=120s`,
	} {
		got, err := authed().RunScript(context.Background(), &labv1.RunScriptRequest{Host: controlPlane, Script: script + "\n"})
		if err != nil || got.ExitCode != 0 {
			t.Errorf("%s: %q, %v", name, got.GetOutput(), err)
		}
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
