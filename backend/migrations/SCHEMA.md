# Database Schema Documentation

## Overview

The Customer Success Tool uses PostgreSQL to store Thena request data with full history tracking.

## Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        thena_requests                           │
├─────────────────────────────────────────────────────────────────┤
│ PK  id                     UUID                                 │
│ UK  thena_id               VARCHAR(255)  ◄─────────┐           │
│     request_id             INTEGER                  │           │
│     event_id               VARCHAR(255)             │           │
│     status                 VARCHAR(50)              │           │
│     sub_status             VARCHAR(255)             │           │
│     sub_status_name        VARCHAR(255)             │           │
│     sub_status_description TEXT                     │           │
│     customer_name          VARCHAR(255)             │           │
│     crm_account_id         VARCHAR(255)             │           │
│     crm_account_name       VARCHAR(255)             │           │
│     channel_id             VARCHAR(255)             │           │
│     channel_name           VARCHAR(255)             │           │
│     description            TEXT                     │           │
│     permalink              TEXT                     │           │
│     thena_url              TEXT                     │           │
│     requestor_*            VARCHAR(255)             │           │
│     assigned_to_*          VARCHAR(255)             │           │
│     assigned_by_*          VARCHAR(255)             │           │
│     sender_*               VARCHAR(255)             │           │
│     reply_count            INTEGER                  │           │
│     first_response_at      VARCHAR(255)             │           │
│     last_customer_*        VARCHAR(255)             │           │
│     last_vendor_*          VARCHAR(255)             │           │
│     created_at             TIMESTAMP WITH TZ        │           │
│     updated_at             TIMESTAMP WITH TZ        │           │
│     received_at            TIMESTAMP WITH TZ        │           │
│     created_at_db          TIMESTAMP WITH TZ        │           │
│     updated_at_db          TIMESTAMP WITH TZ        │           │
└─────────────────────────────────────────────────────┘           │
                                                                   │
                                                                   │
                                                                   │
┌──────────────────────────────────────────────────────────────────┘
│
│   CASCADE DELETE
│
▼
┌─────────────────────────────────────────────────────────────────┐
│                    thena_status_history                         │
├─────────────────────────────────────────────────────────────────┤
│ PK  id                     UUID                                 │
│ FK  thena_id               VARCHAR(255) ───► thena_requests     │
│     request_id             INTEGER                              │
│     from_status            VARCHAR(50)                          │
│     to_status              VARCHAR(50)                          │
│     from_sub_status        VARCHAR(255)                         │
│     to_sub_status          VARCHAR(255)                         │
│     from_sub_status_name   VARCHAR(255)                         │
│     to_sub_status_name     VARCHAR(255)                         │
│     event_id               VARCHAR(255)                         │
│     changed_at             TIMESTAMP WITH TZ                    │
│     created_at_db          TIMESTAMP WITH TZ                    │
│     changed_by_id          VARCHAR(255)                         │
│     changed_by_name        VARCHAR(255)                         │
│     change_reason          TEXT                                 │
│                                                                  │
│ UK  (event_id, to_status, to_sub_status)                        │
└──────────────────────────────────────────────────────────────────┘
```

## Table Descriptions

### `thena_requests`

**Purpose**: Stores the current state of each Thena customer request.

**Key Features**:
- Primary key: `id` (UUID)
- Unique constraint on `thena_id` (Thena's identifier)
- Comprehensive audit trail with multiple timestamps
- Full denormalization of people data (requestor, assigned_to, assigned_by, sender)

**Indexes**:
- `thena_id` - For lookups by Thena ID
- `request_id` - For numeric request ID queries
- `status` - For filtering by current status
- `customer_name` - For customer-specific queries
- `event_id` - For deduplication checks
- `updated_at` - For time-based queries

### `thena_status_history`

**Purpose**: Historical record of all status transitions for complete audit trail.

**Key Features**:
- Foreign key to `thena_requests.thena_id` with CASCADE DELETE
- Tracks both status and substatus changes
- Idempotency via unique constraint on `(event_id, to_status, to_sub_status)`
- Preserves `changed_at` timestamp from Thena events

**Indexes**:
- `thena_id` - For getting history of a specific request
- `request_id` - For convenience lookups
- `event_id` - For event-based queries
- `changed_at` - For time-based analysis
- `to_status` - For status transition analytics

**Unique Constraint**:
```sql
UNIQUE (event_id, to_status, to_sub_status)
```
This prevents duplicate history entries for the same event.

### `schema_migrations`

**Purpose**: Track which migrations have been applied.

**Structure**:
- `version` (PK) - Migration version number
- `applied_at` - Timestamp when migration was applied
- `description` - Human-readable description

## Data Flow

```
Thena Webhook
      ↓
thena-sync service
      ↓
POST /internal/thena/events
      ↓
Backend (Go)
      ↓
┌─────────────────────────┐
│ 1. Check if request     │
│    exists (thena_id)    │
└───────┬─────────────────┘
        │
        ├─── NEW REQUEST ───────────┐
        │                           ↓
        │                   INSERT INTO thena_requests
        │                           ↓
        │                   INSERT INTO thena_status_history
        │                   (from_status = NULL)
        │
        └─── EXISTING REQUEST ──────┐
                                     ↓
                            Compare status/substatus
                                     ↓
                            ┌────────┴────────┐
                            │                 │
                         CHANGED           NO CHANGE
                            │                 │
                            ↓                 ↓
                    UPDATE thena_requests    Skip
                            ↓
                    INSERT INTO thena_status_history
                    (with from/to values)
```

## Status Tracking Logic

When a webhook arrives:

1. **Extract Data**: Parse webhook payload for status information
2. **Upsert Request**:
   - If `thena_id` doesn't exist: INSERT new request
   - If `thena_id` exists: UPDATE existing request
3. **Record History**:
   - For new requests: Record initial status (from_status = NULL)
   - For updates: Only insert if status/substatus changed
   - Use `event_id` to prevent duplicates

## Query Examples

### Get current status of all requests
```sql
SELECT thena_id, request_id, customer_name, status, sub_status_name
FROM thena_requests
ORDER BY updated_at DESC;
```

### Get complete history of a request
```sql
SELECT
    changed_at,
    from_status,
    to_status,
    from_sub_status_name,
    to_sub_status_name,
    changed_by_name
FROM thena_status_history
WHERE thena_id = '696787c382617dc7002dcdb0'
ORDER BY changed_at ASC;
```

### Find requests that changed to "ONHOLD" today
```sql
SELECT DISTINCT r.*
FROM thena_requests r
JOIN thena_status_history h ON r.thena_id = h.thena_id
WHERE h.to_status = 'ONHOLD'
  AND h.changed_at >= CURRENT_DATE;
```

### Count status transitions by customer
```sql
SELECT
    r.customer_name,
    h.to_status,
    COUNT(*) as transition_count
FROM thena_status_history h
JOIN thena_requests r ON h.thena_id = r.thena_id
GROUP BY r.customer_name, h.to_status
ORDER BY r.customer_name, transition_count DESC;
```

## Maintenance

### Viewing Schema
```bash
kubectl exec -it deployment/postgres -n <namespace> -- psql -U postgres -d customer_success
\dt                          # List tables
\d thena_requests            # Describe table
\d+ thena_status_history     # Detailed table info
```

### Backup
```bash
kubectl exec deployment/postgres -n <namespace> -- pg_dump -U postgres customer_success > backup.sql
```

### Restore
```bash
kubectl exec -i deployment/postgres -n <namespace> -- psql -U postgres -d customer_success < backup.sql
```
