#!/bin/bash
# Agent entrypoint script for auto-connecting to the Edictflow server
# Usage: AGENT_EMAIL=agent1@test.local AGENT_PASSWORD=Test1234 ./agent-entrypoint.sh

set -e

# Configuration from environment
AGENT_EMAIL="${AGENT_EMAIL:-agent1@test.local}"
AGENT_PASSWORD="${AGENT_PASSWORD:-Test1234}"
API_SERVER="${EDICTFLOW_API_SERVER:-http://master:8080}"
WS_SERVER="${EDICTFLOW_SERVER:-http://worker:8081}"
POLL_INTERVAL="${POLL_INTERVAL:-500ms}"
MAX_RETRIES="${MAX_RETRIES:-30}"
RETRY_DELAY="${RETRY_DELAY:-2}"

echo "=========================================="
echo "Edictflow Agent Auto-Connect"
echo "=========================================="
echo "Agent Email: $AGENT_EMAIL"
echo "API Server:  $API_SERVER"
echo "WS Server:   $WS_SERVER"
echo "=========================================="

# Wait for API server to be ready
echo "Waiting for API server to be ready..."
retries=0
until curl -sf "${API_SERVER}/health" > /dev/null 2>&1; do
    retries=$((retries + 1))
    if [ $retries -ge $MAX_RETRIES ]; then
        echo "ERROR: API server not ready after $MAX_RETRIES attempts"
        exit 1
    fi
    echo "  Attempt $retries/$MAX_RETRIES - API server not ready, waiting ${RETRY_DELAY}s..."
    sleep $RETRY_DELAY
done
echo "API server is ready!"

# Wait for WebSocket server to be ready (check via master's /health for simplicity)
echo "Waiting for WebSocket server to be ready..."
retries=0
# Worker doesn't have a health endpoint, so we try to establish a connection
# For now, just wait a bit after master is ready
sleep 2
echo "WebSocket server should be ready (assuming it starts with master)"

# Login to the server
echo "Logging in as $AGENT_EMAIL..."
if edictflow login \
    --server "$API_SERVER" \
    --ws-server "$WS_SERVER" \
    --email "$AGENT_EMAIL" \
    --password "$AGENT_PASSWORD"; then
    echo "Login successful!"
else
    echo "ERROR: Login failed"
    exit 1
fi

# Create a sample project directory if it doesn't exist
PROJECT_DIR="/home/agent/projects/sample-project"
if [ ! -d "$PROJECT_DIR" ]; then
    echo "Creating sample project directory..."
    mkdir -p "$PROJECT_DIR"
    cd "$PROJECT_DIR"
    git init
    echo "# Sample Project" > README.md
    git add README.md
    git commit -m "Initial commit"
    echo "Sample project created at $PROJECT_DIR"
fi

# Start the daemon in the foreground
echo "Starting Edictflow daemon..."
echo "  Server: $WS_SERVER"
echo "  Poll Interval: $POLL_INTERVAL"
echo "=========================================="

exec edictflow start \
    --server "$WS_SERVER" \
    --foreground \
    --poll-interval "$POLL_INTERVAL"
