FROM alpine/helm:3.22.0 AS charts
WORKDIR /charts
RUN helm create web \
 && sed -i 's/^version:.*/version: 1.0.0/; s/^appVersion:.*/appVersion: "1.26-alpine"/' web/Chart.yaml \
 && helm package web --destination repo \
 && sed -i 's/^version:.*/version: 1.1.0/; s/^appVersion:.*/appVersion: "1.27-alpine"/' web/Chart.yaml \
 && sed -i 's/^replicaCount:.*/replicaCount: 2/' web/values.yaml \
 && helm package web --destination repo \
 && helm repo index repo --url http://cka-sim-charts

FROM nginx:1.27-alpine
COPY --from=charts /charts/repo /usr/share/nginx/html
