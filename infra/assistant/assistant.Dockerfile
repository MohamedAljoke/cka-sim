# The AI tutor: the chat loop, the MCP tools and the page's Claude Code terminal. It reaches the platform's
# TutorService and answers its Chat, and runs with no AI until it has a Gemini API key or a Vertex AI project.
# Claude Code needs no key: each student signs in with their own Claude account in the terminal.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS go
ARG TARGETOS TARGETARCH
WORKDIR /src/assistant
COPY apps/go-shared /src/go-shared
COPY apps/assistant/go.mod apps/assistant/go.sum ./
RUN GOWORK=off go mod download
COPY apps/assistant .
RUN GOWORK=off CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /cka-sim-assistant ./cmd/assistant

FROM alpine:3.22
ARG CLAUDE_CODE_VERSION=2.1.316
# npm picks Claude Code's musl build; it needs libgcc and libstdc++.
RUN apk add --no-cache nodejs npm libgcc libstdc++ \
  && npm install -g --no-fund --no-audit @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION} \
  && npm cache clean --force && claude --version
COPY --from=go /cka-sim-assistant /usr/local/bin/
ENV CKA_SIM_ASSISTANT_ADDR=:7373 CKA_SIM_ASSISTANT_STATE_DIR=/var/lib/cka-sim-assistant DISABLE_AUTOUPDATER=1
ENTRYPOINT ["cka-sim-assistant"]
