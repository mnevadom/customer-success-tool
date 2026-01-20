-- Migration: 002_feature_requests.sql
-- Description: Create feature_requests table for tracking customer feature requests
-- Date: 2026-01-15

-- ============================================================================
-- Table: feature_requests
-- Description: Track feature requests from customers with Jira integration
-- ============================================================================

CREATE TABLE IF NOT EXISTS feature_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Feature request details
    name VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,

    -- Customer information
    customer_name VARCHAR(255) NOT NULL,
    customers_number_of_requests INTEGER DEFAULT 1,

    -- Jira integration
    jira_link TEXT,
    jira_key VARCHAR(100),

    -- Status tracking
    status VARCHAR(50) DEFAULT 'pending',
    priority VARCHAR(50),

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),

    -- Additional fields for future use
    tags TEXT[],
    estimated_effort VARCHAR(50),
    target_release VARCHAR(100)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_feature_requests_customer ON feature_requests(customer_name);
CREATE INDEX IF NOT EXISTS idx_feature_requests_status ON feature_requests(status);
CREATE INDEX IF NOT EXISTS idx_feature_requests_jira_key ON feature_requests(jira_key);
CREATE INDEX IF NOT EXISTS idx_feature_requests_created_at ON feature_requests(created_at);

-- Record this migration
INSERT INTO schema_migrations (version, description)
VALUES ('002', 'Feature requests table for tracking customer feature requests')
ON CONFLICT (version) DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE feature_requests IS 'Track feature requests from customers with Jira integration';
COMMENT ON COLUMN feature_requests.customers_number_of_requests IS 'Number of customers requesting this feature';
COMMENT ON COLUMN feature_requests.jira_link IS 'Full URL to Jira ticket';
COMMENT ON COLUMN feature_requests.jira_key IS 'Jira ticket key (e.g., PROJ-123)';
