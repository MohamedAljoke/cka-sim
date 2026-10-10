# One exam VM: dockerd plus the cka-sim CLI, which builds the cluster. Built from apps/cka-sim.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cka-sim ./cmd/cli

FROM docker:27-dind
RUN apk add --no-cache bash e2fsprogs
COPY --from=build /cka-sim /usr/local/bin/cka-sim
COPY deploy/fly/lab-start.sh /lab-start.sh
COPY deploy/fly/lab-build.sh /usr/local/bin/lab-build
# dockerd listens on plain TCP, so no certificates.
ENV DOCKER_TLS_CERTDIR=""
ENTRYPOINT ["/lab-start.sh"]
