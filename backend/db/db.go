package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection and runs migrations
func InitDB() error {
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "customer_success")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("🔌 Connecting to database: %s:%s/%s", dbHost, dbPort, dbName)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connection established")

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations runs database migrations
func runMigrations() error {
	log.Println("📦 Running database migrations...")

	// Create schema_migrations table if it doesn't exist
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			description TEXT,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Migration 001: Initial schema
	if err := applyMigration("001", "Initial schema for Thena requests", migration001); err != nil {
		return err
	}

	// Migration 002: Feature requests
	if err := applyMigration("002", "Feature requests table", migration002); err != nil {
		return err
	}

	log.Println("✅ All migrations applied successfully")
	return nil
}

// applyMigration applies a single migration if not already applied
func applyMigration(version, description, sql string) error {
	// Check if migration already applied
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration %s: %w", version, err)
	}

	if count > 0 {
		log.Printf("⏭️  Migration %s already applied, skipping", version)
		return nil
	}

	// Apply migration
	log.Printf("⚙️  Applying migration %s: %s", version, description)
	if _, err := DB.Exec(sql); err != nil {
		return fmt.Errorf("failed to apply migration %s: %w", version, err)
	}

	// Record migration
	_, err = DB.Exec("INSERT INTO schema_migrations (version, description) VALUES ($1, $2)", version, description)
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", version, err)
	}

	log.Printf("✅ Migration %s applied successfully", version)
	return nil
}

const migration001 = `
-- Migration: 001_initial_schema.sql
-- Description: Create initial schema for Thena requests

CREATE TABLE IF NOT EXISTS thena_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thena_id VARCHAR(255) UNIQUE NOT NULL,
    request_id INTEGER,
    event_id VARCHAR(255),
    status VARCHAR(100),
    sub_status VARCHAR(255),
    sub_status_name VARCHAR(255),
    sub_status_description TEXT,
    customer_name VARCHAR(255),
    crm_account_id VARCHAR(255),
    crm_account_name VARCHAR(255),
    channel_id VARCHAR(255),
    channel_name VARCHAR(255),
    description TEXT,
    permalink TEXT,
    thena_url TEXT,
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
    reply_count INTEGER DEFAULT 0,
    first_response_at TIMESTAMP WITH TIME ZONE,
    last_customer_message_at TIMESTAMP WITH TIME ZONE,
    last_vendor_message_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at_db TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    received_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_thena_requests_thena_id ON thena_requests(thena_id);
CREATE INDEX IF NOT EXISTS idx_thena_requests_request_id ON thena_requests(request_id);
CREATE INDEX IF NOT EXISTS idx_thena_requests_customer_name ON thena_requests(customer_name);
CREATE INDEX IF NOT EXISTS idx_thena_requests_status ON thena_requests(status);
CREATE INDEX IF NOT EXISTS idx_thena_requests_sub_status ON thena_requests(sub_status);
CREATE INDEX IF NOT EXISTS idx_thena_requests_created_at ON thena_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_thena_requests_updated_at ON thena_requests(updated_at);

CREATE TABLE IF NOT EXISTS thena_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thena_id VARCHAR(255) NOT NULL,
    request_id INTEGER,
    event_id VARCHAR(255),
    from_status VARCHAR(100),
    to_status VARCHAR(100),
    from_sub_status VARCHAR(255),
    to_sub_status VARCHAR(255),
    from_sub_status_name VARCHAR(255),
    to_sub_status_name VARCHAR(255),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(event_id, to_status, to_sub_status)
);

CREATE INDEX IF NOT EXISTS idx_thena_status_history_thena_id ON thena_status_history(thena_id);
CREATE INDEX IF NOT EXISTS idx_thena_status_history_request_id ON thena_status_history(request_id);
CREATE INDEX IF NOT EXISTS idx_thena_status_history_changed_at ON thena_status_history(changed_at);
`

const migration002 = `
-- Migration: 002_feature_requests.sql
-- Description: Create feature_requests table for tracking customer feature requests

CREATE TABLE IF NOT EXISTS feature_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customers_number_of_requests INTEGER DEFAULT 1,
    jira_link TEXT,
    jira_key VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending',
    priority VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    tags TEXT[],
    estimated_effort VARCHAR(50),
    target_release VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_feature_requests_customer ON feature_requests(customer_name);
CREATE INDEX IF NOT EXISTS idx_feature_requests_status ON feature_requests(status);
CREATE INDEX IF NOT EXISTS idx_feature_requests_jira_key ON feature_requests(jira_key);
CREATE INDEX IF NOT EXISTS idx_feature_requests_created_at ON feature_requests(created_at);
`

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
