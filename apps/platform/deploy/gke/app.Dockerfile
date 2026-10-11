# The platform: the page, its server with the GKE provider, and Caddy in front with HTTPS.
# No Docker: every lab is claimed from the pool and reached through its agent. Built from apps/.
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web .
RUN npm run build

FROM golang:1.26-alpine AS go
# The platform's go.mod points at ../go-shared, so the two sit side by side as in the repo.
WORKDIR /src/platform
COPY go-shared /src/go-shared
COPY platform/go.mod platform/go.sum ./
RUN go mod download
COPY platform .
RUN CGO_ENABLED=0 go build -o /cka-sim-server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache caddy
COPY --from=go /cka-sim-server /usr/local/bin/
COPY --from=web /web/dist /srv
COPY platform/deploy/gke/Caddyfile /etc/caddy/Caddyfile
COPY platform/deploy/gke/app-start.sh /app-start.sh
ENTRYPOINT ["/app-start.sh"]
