# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies for CGO (required for SQLite)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files for both agent and pkg
COPY agent/go.mod agent/go.sum ./agent/
COPY pkg/go.mod ./pkg/

# Download dependencies
WORKDIR /app/agent
RUN go mod download

# Copy source
WORKDIR /app
COPY agent/ ./agent/
COPY pkg/ ./pkg/

# Build agent with CGO enabled (required for SQLite)
WORKDIR /app/agent
RUN CGO_ENABLED=1 GOOS=linux go build -o /edictflow-agent ./cmd/agent

# Runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates sqlite sqlite-libs bash curl git nodejs npm

WORKDIR /home/agent

# Create agent user
RUN adduser -D -h /home/agent agent && \
    mkdir -p /home/agent/.edictflow /home/agent/projects /home/agent/.claude && \
    chown -R agent:agent /home/agent

# Copy binary
COPY --from=builder /edictflow-agent /usr/local/bin/edictflow

# Install Claude Code globally
RUN npm install -g @anthropic-ai/claude-code

USER agent

# Configure git for Claude Code
RUN git config --global user.email "agent@test.local" && \
    git config --global user.name "Test Agent"

# Default environment
ENV HOME=/home/agent
# WebSocket server for daemon connections
ENV EDICTFLOW_SERVER=http://worker:8081
# API server for login/auth
ENV EDICTFLOW_API_SERVER=http://master:8080

# Keep container running for interactive use
# Users can exec into the container and run:
#   edictflow login --server $EDICTFLOW_API_SERVER
#   edictflow start --server $EDICTFLOW_SERVER --foreground --poll-interval 500ms
#   claude - Start Claude Code CLI
CMD ["tail", "-f", "/dev/null"]
