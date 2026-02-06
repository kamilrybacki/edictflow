#!/bin/bash
# Run agent-server E2E tests against the local containerized stack
#
# Prerequisites:
#   1. Docker compose stack must be running (docker compose up -d)
#   2. Database must be seeded with test data
#
# Usage:
#   ./run-agent-server-tests.sh          # Run all tests
#   ./run-agent-server-tests.sh quick    # Run only quick tests (no stress)
#   ./run-agent-server-tests.sh stress   # Run only stress tests
#   ./run-agent-server-tests.sh seed     # Seed database only

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Configuration - can be overridden by environment variables
MASTER_URL="${EDICTFLOW_MASTER_URL:-http://localhost:18080}"
WORKER_URL="${EDICTFLOW_WORKER_URL:-ws://localhost:18081}"
DB_URL="${EDICTFLOW_DB_URL:-postgres://edictflow:edictflow@localhost:15432/edictflow?sslmode=disable}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Edictflow Agent-Server E2E Tests"
echo "=========================================="
echo "Master URL: $MASTER_URL"
echo "Worker URL: $WORKER_URL"
echo "=========================================="

# Check if the stack is running
check_stack() {
    echo -n "Checking if master is reachable... "
    if curl -sf "$MASTER_URL/health" > /dev/null 2>&1; then
        echo -e "${GREEN}OK${NC}"
    else
        echo -e "${RED}FAILED${NC}"
        echo "Error: Master server is not reachable at $MASTER_URL"
        echo "Make sure the docker compose stack is running:"
        echo "  docker compose up -d"
        exit 1
    fi

    echo -n "Checking if worker is reachable... "
    WORKER_HTTP_URL=$(echo "$WORKER_URL" | sed 's/^ws/http/')
    if curl -sf "${WORKER_HTTP_URL}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}OK${NC}"
    else
        echo -e "${RED}FAILED${NC}"
        echo "Error: Worker server is not reachable"
        echo "Make sure the docker compose stack is running:"
        echo "  docker compose up -d"
        exit 1
    fi
}

# Seed the database
seed_database() {
    echo "Seeding database with test data..."
    if [ -f "$PROJECT_ROOT/tests/infrastructure/seed-data.sql" ]; then
        cat "$PROJECT_ROOT/tests/infrastructure/seed-data.sql" | \
            docker compose exec -T db psql -U edictflow 2>/dev/null || {
            echo -e "${YELLOW}Warning: Could not seed database via docker. Trying direct connection...${NC}"
            psql "$DB_URL" < "$PROJECT_ROOT/tests/infrastructure/seed-data.sql" 2>/dev/null || {
                echo -e "${RED}Error: Could not seed database${NC}"
                exit 1
            }
        }
        echo -e "${GREEN}Database seeded successfully${NC}"
    else
        echo -e "${RED}Error: Seed data file not found${NC}"
        exit 1
    fi
}

# Check if test users exist
check_test_users() {
    echo -n "Checking if test users exist... "
    USER_COUNT=$(docker compose exec -T db psql -U edictflow -t -c "SELECT COUNT(*) FROM users WHERE email LIKE '%@test.local';" 2>/dev/null | tr -d ' \n' || echo "0")
    if [ "$USER_COUNT" -gt "0" ]; then
        echo -e "${GREEN}OK (found $USER_COUNT test users)${NC}"
        return 0
    else
        echo -e "${YELLOW}Not found${NC}"
        return 1
    fi
}

# Run tests
run_tests() {
    local test_pattern="$1"
    local test_name="$2"

    echo ""
    echo "Running $test_name tests..."
    echo "----------------------------------------"

    cd "$SCRIPT_DIR"

    EDICTFLOW_MASTER_URL="$MASTER_URL" \
    EDICTFLOW_WORKER_URL="$WORKER_URL" \
    EDICTFLOW_DB_URL="$DB_URL" \
    go test -v -run "$test_pattern" -timeout 300s 2>&1 | while read line; do
        if [[ "$line" =~ "--- PASS" ]]; then
            echo -e "${GREEN}$line${NC}"
        elif [[ "$line" =~ "--- FAIL" ]]; then
            echo -e "${RED}$line${NC}"
        elif [[ "$line" =~ "--- SKIP" ]]; then
            echo -e "${YELLOW}$line${NC}"
        else
            echo "$line"
        fi
    done

    # Get exit status from pipe
    return ${PIPESTATUS[0]}
}

# Main
case "${1:-all}" in
    seed)
        check_stack
        seed_database
        ;;
    quick)
        check_stack
        check_test_users || seed_database
        run_tests "TestAgentServerStack|TestAgentAuthentication|TestAgentWebSocketConnection|TestAgentMultiConnection|TestAPIEndpoints" "quick"
        ;;
    stress)
        check_stack
        check_test_users || seed_database
        run_tests "TestAgentStress" "stress"
        ;;
    all|"")
        check_stack
        check_test_users || seed_database
        run_tests "TestAgent|TestChange|TestRuleUpdate|TestException|TestAPI" "all"
        ;;
    *)
        echo "Usage: $0 [all|quick|stress|seed]"
        echo ""
        echo "Commands:"
        echo "  all     Run all agent-server tests (default)"
        echo "  quick   Run quick tests only (no stress tests)"
        echo "  stress  Run stress tests only"
        echo "  seed    Seed the database only"
        exit 1
        ;;
esac

echo ""
echo "=========================================="
echo "Tests completed!"
echo "=========================================="
