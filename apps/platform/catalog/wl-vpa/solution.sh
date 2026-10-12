kubectl apply -f - <<'YAML'
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: api
  namespace: wl-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api
  updatePolicy:
    updateMode: "Off"
YAML
for _ in $(seq 60); do
  cpu=$(kubectl -n wl-vpa get vpa api -o jsonpath='{.status.recommendation.containerRecommendations[0].target.cpu}')
  [ -n "$cpu" ] && break
  sleep 3
done
mem=$(kubectl -n wl-vpa get vpa api -o jsonpath='{.status.recommendation.containerRecommendations[0].target.memory}')
echo "cpu=$cpu memory=$mem" > /opt/course/wl-vpa/target.txt
