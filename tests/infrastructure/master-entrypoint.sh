#!/bin/sh
set -e

echo "Running database migrations..."
./migrate -path ./migrations -database "$DATABASE_URL" up

# Check if seed data file exists and run it
if [ -f /seed-data.sql ]; then
    echo "Seeding test data..."
    # Extract connection params from DATABASE_URL
    # Format: postgres://user:pass@host:port/dbname?sslmode=disable
    DB_HOST=$(echo $DATABASE_URL | sed -n 's|.*@\([^:]*\):.*|\1|p')
    DB_PORT=$(echo $DATABASE_URL | sed -n 's|.*:\([0-9]*\)/.*|\1|p')
    DB_NAME=$(echo $DATABASE_URL | sed -n 's|.*/\([^?]*\).*|\1|p')
    DB_USER=$(echo $DATABASE_URL | sed -n 's|.*://\([^:]*\):.*|\1|p')

    # Use PGPASSWORD from DATABASE_URL
    export PGPASSWORD=$(echo $DATABASE_URL | sed -n 's|.*://[^:]*:\([^@]*\)@.*|\1|p')

    # Check if psql is available, if not skip seeding
    if command -v psql > /dev/null 2>&1; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f /seed-data.sql || echo "Seed data already applied or failed"
    else
        echo "Warning: psql not available, skipping seed data"
    fi
fi

echo "Starting master server..."
exec ./master
