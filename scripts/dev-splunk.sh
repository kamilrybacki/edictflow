#!/bin/bash
# Start local development stack with Splunk observability
#
# Usage: ./scripts/dev-splunk.sh [up|down|logs|status]
#
# Access points after startup (using non-standard ports to avoid conflicts):
#   - Splunk Web UI: http://localhost:18000
#     Username: admin
#     Password: DevPassword123!
#   - Splunk HEC: http://localhost:18088
#   - API Master: http://localhost:18080
#   - Worker WebSocket: ws://localhost:18081
#   - Web UI: http://localhost:13000
#   - Redis: localhost:16379
#   - PostgreSQL: localhost:15432

set -e

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.dev.yml"

case "${1:-up}" in
  up)
    echo "Starting Claudeception with Splunk observability..."
    echo ""
    docker compose $COMPOSE_FILES up -d
    echo ""
    echo "Waiting for Splunk to be ready (this may take 1-2 minutes on first start)..."
    echo ""

    # Wait for Splunk to be healthy
    timeout=180
    elapsed=0
    while [ $elapsed -lt $timeout ]; do
      if docker compose $COMPOSE_FILES ps splunk | grep -q "healthy"; then
        break
      fi
      sleep 5
      elapsed=$((elapsed + 5))
      echo "  Waiting... ($elapsed seconds)"
    done

    echo ""
    echo "=========================================="
    echo "  Claudeception Dev Stack Ready!"
    echo "=========================================="
    echo ""
    echo "  Splunk Web UI:    http://localhost:18000"
    echo "    Username:       admin"
    echo "    Password:       DevPassword123!"
    echo ""
    echo "  Splunk HEC:       http://localhost:18088"
    echo "    Token:          claudeception-dev-token"
    echo ""
    echo "  API Master:       http://localhost:18080"
    echo "  Worker WebSocket: ws://localhost:18081"
    echo "  Web UI:           http://localhost:13000"
    echo ""
    echo "  Redis:            localhost:16379"
    echo "  PostgreSQL:       localhost:15432"
    echo ""
    echo "  To search metrics in Splunk:"
    echo "    index=claudeception | head 100"
    echo ""
    echo "  Useful Splunk searches:"
    echo "    index=claudeception type=api_request | stats count by path"
    echo "    index=claudeception type=agent_connection | timechart count by action"
    echo "    index=claudeception type=health_check | stats latest(status) by component"
    echo "    index=claudeception type=hub_stats | timechart avg(agents) avg(teams)"
    echo ""
    ;;

  down)
    echo "Stopping Claudeception dev stack..."
    docker compose $COMPOSE_FILES down
    ;;

  logs)
    docker compose $COMPOSE_FILES logs -f "${2:-}"
    ;;

  status)
    docker compose $COMPOSE_FILES ps
    ;;

  restart)
    echo "Restarting Claudeception dev stack..."
    docker compose $COMPOSE_FILES restart
    ;;

  clean)
    echo "Stopping and removing all data (including Splunk data)..."
    docker compose $COMPOSE_FILES down -v
    ;;

  *)
    echo "Usage: $0 [up|down|logs|status|restart|clean]"
    echo ""
    echo "Commands:"
    echo "  up      - Start the dev stack with Splunk"
    echo "  down    - Stop the dev stack"
    echo "  logs    - Follow logs (optionally specify service: logs splunk)"
    echo "  status  - Show container status"
    echo "  restart - Restart all services"
    echo "  clean   - Stop and remove all data volumes"
    exit 1
    ;;
esac
