# One lab: dockerd, the lab CLI that builds the cluster, and the agent the platform reaches it through. The
# same image runs as a container on a laptop (make up) and as a pod on its own GKE node. Built from apps/lab.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /lab ./cmd/lab && CGO_ENABLED=0 go build -o /cka-sim-agent ./cmd/agent

FROM docker:27-dind
RUN apk add --no-cache bash
COPY --from=build /lab /cka-sim-agent /usr/local/bin/
COPY deploy/lab-start.sh /lab-start.sh
ENV DOCKER_TLS_CERTDIR=""
ENTRYPOINT ["/lab-start.sh"]
