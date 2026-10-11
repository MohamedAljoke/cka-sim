FROM golang:1.26-alpine AS build
WORKDIR /src/auth
COPY go-shared /src/go-shared
COPY auth/go.mod auth/go.sum ./
RUN go mod download
COPY auth .
RUN CGO_ENABLED=0 go build -o /cka-sim-auth ./cmd/server && \
    CGO_ENABLED=0 go build -o /sql-migrate github.com/rubenv/sql-migrate/sql-migrate

FROM alpine:3.22
COPY --from=build /cka-sim-auth /sql-migrate /usr/local/bin/
COPY auth/scripts/db /db
WORKDIR /db
ENTRYPOINT ["cka-sim-auth"]
