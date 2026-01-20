#!/bin/sh
# Migration runner script for Customer Success Tool database
# This script applies all pending migrations to the PostgreSQL database

set -e

echo "🗄️  Starting database migrations..."

# Database connection parameters
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-customer_success}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

# Construct connection string
DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

echo "📡 Connecting to database: ${DB_HOST}:${DB_PORT}/${DB_NAME}"

# Wait for database to be ready
until psql "$DB_URL" -c '\q' 2>/dev/null; do
  echo "⏳ Waiting for PostgreSQL to be ready..."
  sleep 2
done

echo "✅ Database is ready"

# Run migrations in order
MIGRATIONS_DIR="/app/migrations"
if [ ! -d "$MIGRATIONS_DIR" ]; then
  MIGRATIONS_DIR="./migrations"
fi

echo "📂 Looking for migrations in: $MIGRATIONS_DIR"

for migration_file in "$MIGRATIONS_DIR"/[0-9]*.sql; do
  if [ -f "$migration_file" ]; then
    migration_name=$(basename "$migration_file")
    echo ""
    echo "🔄 Applying migration: $migration_name"

    if psql "$DB_URL" -f "$migration_file"; then
      echo "✅ Migration $migration_name applied successfully"
    else
      echo "❌ Migration $migration_name failed"
      exit 1
    fi
  fi
done

echo ""
echo "✅ All migrations completed successfully!"

# Show applied migrations
echo ""
echo "📋 Applied migrations:"
psql "$DB_URL" -c "SELECT version, applied_at, description FROM schema_migrations ORDER BY version;" 2>/dev/null || echo "   (schema_migrations table not yet created)"

echo ""
echo "🎉 Database is up to date!"
