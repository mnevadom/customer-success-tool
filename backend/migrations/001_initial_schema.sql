-- Migration: 001_initial_schema.sql
-- Description: Create initial tables for Thena requests and status history
-- Date: 2026-01-15

-- ============================================================================
-- Table: thena_requests
-- Description: Main table storing current state of Thena customer requests
-- ============================================================================

CREATE TABLE IF NOT EXISTS thena_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Thena identifiers
    thena_id VARCHAR(255) UNIQUE NOT NULL,
    request_id INTEGER NOT NULL,
    event_id VARCHAR(255) NOT NULL,

    -- Current status
    status VARCHAR(50) NOT NULL,
    sub_status VARCHAR(255),
    sub_status_name VARCHAR(255),
    sub_status_description TEXT,

    -- Customer information
    customer_name VARCHAR(255) NOT NULL,
    crm_account_id VARCHAR(255),
    crm_account_name VARCHAR(255),

    -- Channel information
    channel_id VARCHAR(255),
    channel_name VARCHAR(255),

    -- Request details
    description TEXT,
    permalink TEXT,
    thena_url TEXT,

    -- People involved
    requestor_id VARCHAR(255),
    requestor_name VARCHAR(255),
    requestor_email VARCHAR(255),
    requestor_domain VARCHAR(255),

    assigned_to_id VARCHAR(255),
    assigned_to_name VARCHAR(255),
    assigned_to_email VARCHAR(255),
    assigned_to_domain VARCHAR(255),

    assigned_by_id VARCHAR(255),
    assigned_by_name VARCHAR(255),
    assigned_by_email VARCHAR(255),
    assigned_by_domain VARCHAR(255),

    sender_id VARCHAR(255),
    sender_name VARCHAR(255),
    sender_email VARCHAR(255),
    sender_domain VARCHAR(255),

    -- Metrics
    reply_count INTEGER DEFAULT 0,
    first_response_at VARCHAR(255),
    last_customer_message_at VARCHAR(255),
    last_vendor_message_at VARCHAR(255),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Metadata
    created_at_db TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at_db TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_thena_requests_thena_id ON thena_requests(thena_id);
CREATE INDEX IF NOT EXISTS idx_thena_requests_request_id ON thena_requests(request_id);
CREATE INDEX IF NOT EXISTS idx_thena_requests_status ON thena_requests(status);
CREATE INDEX IF NOT EXISTS idx_thena_requests_customer_name ON thena_requests(customer_name);
CREATE INDEX IF NOT EXISTS idx_thena_requests_event_id ON thena_requests(event_id);
CREATE INDEX IF NOT EXISTS idx_thena_requests_updated_at ON thena_requests(updated_at);

-- ============================================================================
-- Table: thena_status_history
-- Description: Track all status changes for Thena requests
-- ============================================================================

CREATE TABLE IF NOT EXISTS thena_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign key to main table
    thena_id VARCHAR(255) NOT NULL REFERENCES thena_requests(thena_id) ON DELETE CASCADE,
    request_id INTEGER NOT NULL,

    -- Status change details
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    from_sub_status VARCHAR(255),
    to_sub_status VARCHAR(255),
    from_sub_status_name VARCHAR(255),
    to_sub_status_name VARCHAR(255),

    -- Event tracking for idempotency
    event_id VARCHAR(255) NOT NULL,

    -- Timestamps
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at_db TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Additional context
    changed_by_id VARCHAR(255),
    changed_by_name VARCHAR(255),
    change_reason TEXT,

    CONSTRAINT unique_event_status_change UNIQUE (event_id, to_status, to_sub_status)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_status_history_thena_id ON thena_status_history(thena_id);
CREATE INDEX IF NOT EXISTS idx_status_history_request_id ON thena_status_history(request_id);
CREATE INDEX IF NOT EXISTS idx_status_history_event_id ON thena_status_history(event_id);
CREATE INDEX IF NOT EXISTS idx_status_history_changed_at ON thena_status_history(changed_at);
CREATE INDEX IF NOT EXISTS idx_status_history_to_status ON thena_status_history(to_status);

-- ============================================================================
-- Table: schema_migrations
-- Description: Track which migrations have been applied
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    description TEXT
);

-- Record this migration
INSERT INTO schema_migrations (version, description)
VALUES ('001', 'Initial schema - thena_requests and thena_status_history tables')
ON CONFLICT (version) DO NOTHING;

-- ============================================================================
-- Comments for documentation
-- ============================================================================

COMMENT ON TABLE thena_requests IS 'Main table storing current state of Thena customer requests';
COMMENT ON TABLE thena_status_history IS 'Historical record of all status changes for Thena requests';
COMMENT ON TABLE schema_migrations IS 'Track applied database migrations';

COMMENT ON COLUMN thena_requests.thena_id IS 'Unique Thena request identifier';
COMMENT ON COLUMN thena_requests.event_id IS 'Latest event ID for deduplication';
COMMENT ON COLUMN thena_status_history.event_id IS 'Event ID that triggered this status change (for idempotency)';
COMMENT ON COLUMN thena_status_history.changed_at IS 'Timestamp from the event, or NOW() if not available';
