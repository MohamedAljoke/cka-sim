FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src/lab
COPY go-shared /src/go-shared
COPY lab/go.mod lab/go.sum ./
RUN GOWORK=off go mod download
COPY lab .
RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-agent ./cmd/agent

FROM docker:28-dind
COPY --from=build /cka-sim-agent /usr/local/bin/
COPY lab/deploy/lab-start.sh /lab-start.sh
# dind wraps dockerd in its own tini, which reaps only when it is PID 1 or a subreaper.
ENV DOCKER_TLS_CERTDIR="" TINI_SUBREAPER=1
ENTRYPOINT ["/lab-start.sh"]
