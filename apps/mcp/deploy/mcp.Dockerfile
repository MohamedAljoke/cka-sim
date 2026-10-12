FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src/mcp
COPY go-shared /src/go-shared
COPY mcp/go.mod mcp/go.sum ./
RUN go mod download
COPY mcp .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-mcp ./cmd/server

FROM alpine:3.22
COPY --from=build /cka-sim-mcp /usr/local/bin/
ENTRYPOINT ["cka-sim-mcp"]
