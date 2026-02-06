#!/bin/bash
# Setup script for test infrastructure
# Runs migrations and seeds test data

set -e

echo "Waiting for database to be ready..."
until PGPASSWORD=edictflow psql -h db -U edictflow -d edictflow -c '\q' 2>/dev/null; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done

echo "Database is ready!"

echo "Running migrations..."
cd /app/server
migrate -path migrations -database "postgres://edictflow:edictflow@db:5432/edictflow?sslmode=disable" up

echo "Seeding test data..."
PGPASSWORD=edictflow psql -h db -U edictflow -d edictflow -f /app/test-infra/seed-data.sql

echo "Setup complete!"
