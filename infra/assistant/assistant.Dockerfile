# The AI tutor: the chat loop and the MCP tools. It reaches the platform's TutorService and answers its Chat, and
# runs with no AI until it has a Gemini API key or a Vertex AI project.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS go
ARG TARGETOS TARGETARCH
WORKDIR /src/assistant
COPY apps/go-shared /src/go-shared
COPY apps/assistant/go.mod apps/assistant/go.sum ./
RUN GOWORK=off go mod download
COPY apps/assistant .
RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-assistant ./cmd/assistant

FROM alpine:3.22
COPY --from=go /cka-sim-assistant /usr/local/bin/
ENV CKA_SIM_ASSISTANT_ADDR=:7373
ENTRYPOINT ["cka-sim-assistant"]
