kubectl is a REST client. Every object has a path built from its group, version, namespace and
name:

```
/api/v1/namespaces/<ns>/pods                       # core group: /api/v1
/apis/apps/v1/namespaces/<ns>/deployments/<name>   # named groups: /apis/<group>/<version>
/apis/storage.k8s.io/v1/storageclasses             # cluster-scoped: no namespaces/<ns>
```

Two easy ways to call it with your kubeconfig's credentials:

```sh
kubectl get --raw /api/v1/namespaces/ar-raw-api/pods | jq '.items[].metadata.name'

kubectl proxy --port=8001 &                         # handles TLS and auth for you
curl -s localhost:8001/apis/apps/v1/namespaces/ar-raw-api/deployments/web | jq .spec.replicas
kill %1
```

`kubectl get pods -v=8` prints the exact URLs kubectl requests, which is the quickest way to learn a
path. Without kubectl, you'd pass a bearer token (`curl -H "Authorization: Bearer $TOKEN" --cacert
ca.crt https://...:6443/...`), which is how apps inside Pods call the API with their ServiceAccount
token.
