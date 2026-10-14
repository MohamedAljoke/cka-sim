# Embeds the Kubernetes docs into the docs database once, as a Job: the docs of DOCS_SHA and the indexer. Built from apps/.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS go
ARG TARGETOS TARGETARCH
WORKDIR /src/platform
COPY go-shared /src/go-shared
COPY platform/go.mod platform/go.sum ./
RUN go mod download
COPY platform .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-docs ./cmd/docs

FROM --platform=$BUILDPLATFORM alpine:3.22 AS website
ARG DOCS_SHA
RUN test -n "$DOCS_SHA" && apk add --no-cache git && \
    git init -q /website && cd /website && \
    git remote add origin https://github.com/kubernetes/website && \
    git sparse-checkout set content/en/docs content/en/examples && \
    git fetch -q --depth 1 --filter=blob:none origin "$DOCS_SHA" && git checkout -q FETCH_HEAD

FROM alpine:3.22
ARG DOCS_SHA
ENV DOCS_SHA=$DOCS_SHA
COPY --from=go /cka-sim-docs /usr/local/bin/
COPY --from=website /website/content /website/content
ENTRYPOINT ["sh", "-c", "exec cka-sim-docs index -dir /website -ref \"$DOCS_SHA\""]
