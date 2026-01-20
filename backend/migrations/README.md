# Database Migrations

This directory contains SQL migration files for the Customer Success Tool database.

## Migration Files

Migrations are numbered sequentially:
- `001_initial_schema.sql` - Initial database schema with thena_requests and thena_status_history tables

## Schema Overview

### `thena_requests` Table
Main table storing the current state of each Thena customer request. Key fields:
- `thena_id` - Unique Thena request identifier (from Thena)
- `request_id` - Numeric request ID (from Thena)
- `event_id` - Latest event ID for deduplication
- `status` - Current status (OPEN, ONHOLD, RESOLVED, etc.)
- `sub_status` - Current substatus ID
- `sub_status_name` - Human-readable substatus (e.g., "Waiting for Customer")
- `customer_name` - Customer name
- Full audit trail with timestamps

### `thena_status_history` Table
Historical record of all status changes. Key fields:
- `thena_id` - Foreign key to thena_requests
- `from_status` / `to_status` - Status transition
- `from_sub_status` / `to_sub_status` - Substatus transition
- `event_id` - Event that triggered the change (for idempotency)
- `changed_at` - When the change occurred
- Unique constraint on `(event_id, to_status, to_sub_status)` prevents duplicate entries

### `schema_migrations` Table
Tracks which migrations have been applied to the database.

## Running Migrations

### Manual Method
```bash
kubectl exec -i deployment/postgres -n <namespace> -- psql -U postgres -d customer_success < migrations/001_initial_schema.sql
```

### From Backend Pod
```bash
psql postgresql://postgres:postgres@postgres:5432/customer_success -f /app/migrations/001_initial_schema.sql
```

## Adding New Migrations

1. Create a new file: `00X_description.sql`
2. Include the migration version in `schema_migrations` table
3. Test the migration locally first
4. Update this README with the migration description

## Rollback Strategy

Currently, migrations are forward-only. For rollback:
1. Create a new migration that reverts changes
2. Never delete or modify existing migration files
