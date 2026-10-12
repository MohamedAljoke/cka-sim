# One lab: dockerd, the lab CLI that builds the cluster, and the agent the platform reaches it through. The
# same image runs as a container on a laptop (make up) and as a pod on its own GKE node. Built from apps/.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH
# The lab's go.mod points at ../go-shared, so the two sit side by side as in the repo.
WORKDIR /src/lab
COPY go-shared /src/go-shared
COPY lab/go.mod lab/go.sum ./
RUN go mod download
COPY lab .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /lab ./cmd/lab && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-agent ./cmd/agent

FROM docker:27-dind
RUN apk add --no-cache bash
COPY --from=build /lab /cka-sim-agent /usr/local/bin/
COPY lab/deploy/lab-start.sh /lab-start.sh
ENV DOCKER_TLS_CERTDIR=""
ENTRYPOINT ["/lab-start.sh"]
