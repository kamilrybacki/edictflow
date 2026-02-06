#!/bin/bash
#
# E2E Test Runner for Edictflow
#
# This script runs E2E tests with proper setup and cleanup.
# Usage:
#   ./run-e2e.sh           # Run all E2E tests
#   ./run-e2e.sh --cleanup-only  # Just cleanup orphaned resources
#

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
E2E_DIR="${PROJECT_ROOT}/e2e"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup orphaned E2E resources
cleanup_orphaned() {
    log_info "Cleaning up orphaned E2E resources..."

    # Remove orphaned networks with e2e- prefix
    for net in $(docker network ls --filter "name=e2e-" -q 2>/dev/null); do
        log_info "Removing orphaned network: $net"
        docker network rm "$net" 2>/dev/null || true
    done

    # Remove orphaned containers with e2e- label or name
    for container in $(docker ps -a --filter "name=e2e-" -q 2>/dev/null); do
        log_info "Removing orphaned container: $container"
        docker rm -f "$container" 2>/dev/null || true
    done

    # Clean up temp directories
    rm -rf /tmp/e2e-workspace-* 2>/dev/null || true
    rm -rf /tmp/e2e-agentdb-* 2>/dev/null || true
    rm -f /tmp/agent-linux-* 2>/dev/null || true

    log_info "Cleanup complete"
}

# Check prerequisites
check_prereqs() {
    log_info "Checking prerequisites..."

    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Verify agent builds
    log_info "Verifying agent builds..."
    (cd "${PROJECT_ROOT}/agent" && go build -o /dev/null ./cmd/agent) || {
        log_error "Agent failed to build"
        exit 1
    }

    # Verify server Dockerfile exists
    if [ ! -f "${PROJECT_ROOT}/server/Dockerfile" ]; then
        log_error "Server Dockerfile not found at ${PROJECT_ROOT}/server/Dockerfile"
        exit 1
    fi

    log_info "Prerequisites check passed"
}

# Run E2E tests
run_tests() {
    log_info "Running E2E tests..."

    cd "$E2E_DIR"

    # Run with verbose output and extended timeout
    go test -v -count=1 -timeout 30m ./... 2>&1 | tee e2e-test.log

    exit_code=${PIPESTATUS[0]}

    if [ $exit_code -eq 0 ]; then
        log_info "E2E tests passed!"
    else
        log_error "E2E tests failed with exit code $exit_code"
        log_info "Check e2e/e2e-test.log for details"
    fi

    return $exit_code
}

# Main
main() {
    case "${1:-}" in
        --cleanup-only)
            cleanup_orphaned
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --cleanup-only  Only cleanup orphaned E2E resources"
            echo "  --help, -h      Show this help message"
            echo ""
            echo "Without options, runs the full E2E test suite."
            ;;
        *)
            check_prereqs
            cleanup_orphaned
            run_tests
            ;;
    esac
}

main "$@"
