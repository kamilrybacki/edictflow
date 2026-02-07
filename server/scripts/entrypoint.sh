#!/bin/sh
set -e

echo "Running database migrations..."

# Wait for database to be ready by checking if we can connect
max_retries=30
retry_count=0
until ./migrate -path ./migrations -database "$DATABASE_URL" version 2>&1 | grep -qE "^[0-9]+$|no migration"; do
  retry_count=$((retry_count + 1))
  if [ $retry_count -ge $max_retries ]; then
    echo "Failed to connect to database after $max_retries attempts"
    exit 1
  fi
  echo "Waiting for database to be ready... (attempt $retry_count/$max_retries)"
  sleep 2
done

echo "Database is ready!"

# Run migrations
./migrate -path ./migrations -database "$DATABASE_URL" up

echo "Migrations complete!"

# Start the application
exec "$@"
