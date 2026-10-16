# The assistant: chat, docs search and the MCP tools, plus the docs schema's migrations. Built from apps/.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src/assistant
COPY go-shared /src/go-shared
COPY assistant/go.mod assistant/go.sum ./
RUN go mod download
COPY assistant .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-assistant ./cmd/server && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /sql-migrate github.com/rubenv/sql-migrate/sql-migrate

FROM alpine:3.22
COPY --from=build /cka-sim-assistant /sql-migrate /usr/local/bin/
COPY assistant/scripts/db /db
ENTRYPOINT ["cka-sim-assistant"]
