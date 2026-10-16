# The platform: the page, its server, and Caddy in front. GKE and infra/local both run it.
# No Docker: every lab is reached through its agent. Built from apps/.
FROM --platform=$BUILDPLATFORM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web .
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS go
ARG TARGETOS TARGETARCH
# The platform's go.mod points at ../go-shared, so the two sit side by side as in the repo.
WORKDIR /src/platform
COPY go-shared /src/go-shared
COPY platform/go.mod platform/go.sum ./
RUN go mod download
COPY platform .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache caddy
COPY --from=go /cka-sim-server /usr/local/bin/
COPY --from=web /web/dist /srv
COPY platform/deploy/Caddyfile /etc/caddy/Caddyfile
COPY platform/deploy/app-start.sh /app-start.sh
ENTRYPOINT ["/app-start.sh"]
