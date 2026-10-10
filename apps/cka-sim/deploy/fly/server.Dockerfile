# The site: the built page and the cka-sim server. Built from apps/.
FROM node:24-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS server
WORKDIR /src
COPY cka-sim/go.mod cka-sim/go.sum ./
RUN go mod download
COPY cka-sim/ ./
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.22
# The script runner shells out to docker exec.
RUN apk add --no-cache docker-cli
COPY --from=web /web/dist /web
COPY --from=server /server /usr/local/bin/server
ENV CKA_SIM_ADDR=0.0.0.0:8080 CKA_SIM_WEB=/web
EXPOSE 8080
CMD ["server"]
