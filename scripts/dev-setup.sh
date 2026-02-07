#!/bin/bash
# dev-setup.sh - Clean setup of the development stack
# This script stops containers, removes volumes, and starts fresh

set -e

echo "=== EdictFlow Development Setup ==="
echo ""

# Navigate to project root (parent of scripts directory)
cd "$(dirname "$0")/.."

echo "1. Stopping all containers..."
docker compose down --remove-orphans 2>/dev/null || true

echo "2. Removing volumes (cleaning database and Redis data)..."
docker compose down -v 2>/dev/null || true

echo "3. Pruning unused Docker resources..."
docker system prune -f --volumes 2>/dev/null || true

echo "4. Building containers..."
docker compose build

echo "5. Starting stack..."
docker compose up -d

echo ""
echo "6. Waiting for services to be ready..."
sleep 5

# Wait for master to be ready
max_retries=30
retry_count=0
until curl -s http://localhost:8080/health > /dev/null 2>&1 || [ $retry_count -ge $max_retries ]; do
  retry_count=$((retry_count + 1))
  echo "   Waiting for master API... (attempt $retry_count/$max_retries)"
  sleep 2
done

if [ $retry_count -ge $max_retries ]; then
  echo "   Master API not responding, checking logs..."
  docker compose logs master --tail 20
else
  echo "   Master API is ready!"
fi

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Services:"
echo "  - Web UI:     http://localhost:3000"
echo "  - Master API: http://localhost:8080"
echo "  - Worker:     http://localhost:8081"
echo ""
echo "Test Accounts (Password: Password123):"
echo "  - admin@example.com    (Admin, Engineering team)"
echo "  - user@example.com     (Member, Engineering team)"
echo "  - designer@example.com (Member, Design team)"
echo ""
echo "Test Agents (Password: Test1234):"
echo "  - alex.rivera@test.local, jordan.kim@test.local (Engineering team)"
echo "  - sarah.chen@test.local                         (Design team)"
echo "  - mike.johnson@test.local, emma.wilson@test.local (Operations team)"
echo ""
echo "Team Invite Codes:"
echo "  - ENGJOIN  (Engineering team)"
echo "  - DESNJOIN (Design team)"
echo "  - OPSJOIN  (Operations team)"
echo ""
echo "To start test agent containers: task up:test"
echo ""
echo "Logs: task logs"
echo "Stop: task down"
echo ""
