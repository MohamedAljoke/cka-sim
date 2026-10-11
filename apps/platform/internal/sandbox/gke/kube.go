package gke

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	serviceAccount = "/var/run/secrets/kubernetes.io/serviceaccount"
	claimsPath     = "/apis/extensions.agents.x-k8s.io/v1beta1/namespaces/%s/sandboxclaims"
	podPath        = "/api/v1/namespaces/%s/pods/%s"
	managedBy      = "app.kubernetes.io/managed-by"
	ours           = "cka-sim"
	accountLabel   = "cka-sim/account"
	ready          = "Ready"
)

type kube struct {
	base      string
	namespace string
	// token is read on every call: Kubernetes rotates service-account tokens.
	token  func() (string, error)
	client *http.Client
}

type object struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}

type claim struct {
	object
	Status struct {
		Conditions []struct {
			Type   string `json:"type"`
			Status string `json:"status"`
			Reason string `json:"reason"`
		} `json:"conditions"`
		Sandbox struct {
			Name string `json:"name"`
		} `json:"sandbox"`
	} `json:"status"`
}

func (c claim) ready() bool {
	for _, cond := range c.Status.Conditions {
		if cond.Type == ready {
			return cond.Status == "True" && c.Status.Sandbox.Name != ""
		}
	}
	return false
}

type pod struct {
	Status struct {
		PodIP string `json:"podIP"`
	} `json:"status"`
}

func inCluster() (kube, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" {
		return kube{}, errors.New("not in a Kubernetes pod: KUBERNETES_SERVICE_HOST is not set")
	}
	ns, err := os.ReadFile(serviceAccount + "/namespace")
	if err != nil {
		return kube{}, fmt.Errorf("read the pod's namespace: %w", err)
	}
	ca, err := os.ReadFile(serviceAccount + "/ca.crt")
	if err != nil {
		return kube{}, fmt.Errorf("read the cluster's CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		return kube{}, errors.New("the cluster's CA holds no certificate")
	}
	return kube{
		base:      "https://" + net.JoinHostPort(host, port),
		namespace: strings.TrimSpace(string(ns)),
		token: func() (string, error) {
			b, err := os.ReadFile(serviceAccount + "/token")
			return strings.TrimSpace(string(b)), err
		},
		client: &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}}},
	}, nil
}

func (k kube) createClaim(ctx context.Context, pool, owner string) (string, error) {
	body := map[string]any{
		"apiVersion": "extensions.agents.x-k8s.io/v1beta1",
		"kind":       "SandboxClaim",
		"metadata":   map[string]any{"generateName": "lab-", "labels": map[string]string{managedBy: ours, accountLabel: owner}},
		"spec":       map[string]any{"warmPoolRef": map[string]string{"name": pool}},
	}
	var created object
	err := k.do(ctx, http.MethodPost, k.claims(), body, &created)
	return created.Metadata.Name, err
}

func (k kube) getClaim(ctx context.Context, name string) (claim, error) {
	var c claim
	err := k.do(ctx, http.MethodGet, k.claims()+"/"+name, nil, &c)
	return c, err
}

func (k kube) ourClaims(ctx context.Context) ([]string, error) {
	var list struct {
		Items []object `json:"items"`
	}
	if err := k.do(ctx, http.MethodGet, k.claims()+"?labelSelector="+url.QueryEscape(managedBy+"="+ours), nil, &list); err != nil {
		return nil, err
	}
	names := make([]string, len(list.Items))
	for i, item := range list.Items {
		names[i] = item.Metadata.Name
	}
	return names, nil
}

func (k kube) deleteClaim(ctx context.Context, name string) error {
	return k.do(ctx, http.MethodDelete, k.claims()+"/"+name, nil, nil)
}

func (k kube) podIP(ctx context.Context, name string) (string, error) {
	var p pod
	if err := k.do(ctx, http.MethodGet, fmt.Sprintf(podPath, k.namespace, name), nil, &p); err != nil {
		return "", err
	}
	if p.Status.PodIP == "" {
		return "", fmt.Errorf("pod %s has no IP yet", name)
	}
	return p.Status.PodIP, nil
}

func (k kube) claims() string { return fmt.Sprintf(claimsPath, k.namespace) }

func (k kube) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, k.base+path, body)
	if err != nil {
		return err
	}
	token, err := k.token()
	if err != nil {
		return fmt.Errorf("read the service-account token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := k.client.Do(req)
	if err != nil {
		return fmt.Errorf("kubernetes %s %s: %w", method, path, err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("kubernetes %s %s: %s: %s", method, path, res.Status, bytes.TrimSpace(msg))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}
