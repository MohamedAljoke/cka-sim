# The page, its server, and Caddy in front, which serves the page and sends /api and /ws to the server.
# No Docker: the lab is reached through its agent.
FROM --platform=$BUILDPLATFORM node:22-alpine AS web
WORKDIR /web
COPY apps/web/package.json apps/web/package-lock.json ./
RUN npm ci
COPY apps/web .
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS go
ARG TARGETOS TARGETARCH
WORKDIR /src/platform
COPY apps/go-shared /src/go-shared
COPY apps/platform/go.mod apps/platform/go.sum ./
RUN GOWORK=off go mod download
COPY apps/platform .
RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache caddy
COPY --from=go /cka-sim-server /usr/local/bin/
COPY --from=web /web/dist /srv
COPY infra/app/Caddyfile /etc/caddy/Caddyfile
COPY infra/app/app-start.sh /app-start.sh
ENV CKA_SIM_PLATFORM_ADDR=127.0.0.1:7070
ENTRYPOINT ["/app-start.sh"]
