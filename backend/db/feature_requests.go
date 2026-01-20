package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

// FeatureRequestDB represents a feature request in the database
type FeatureRequestDB struct {
	ID                        string
	Name                      string
	Description               string
	CustomerName              string
	CustomersNumberOfRequests int
	JiraLink                  sql.NullString
	JiraKey                   sql.NullString
	Status                    string
	Priority                  sql.NullString
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	CreatedBy                 sql.NullString
	Tags                      pq.StringArray
	EstimatedEffort           sql.NullString
	TargetRelease             sql.NullString
}

// CreateFeatureRequest inserts a new feature request into the database
func CreateFeatureRequest(req FeatureRequestDB) (string, error) {
	var id string
	err := DB.QueryRow(`
		INSERT INTO feature_requests (
			name, description, customer_name, customers_number_of_requests,
			jira_link, jira_key, status, priority, created_by, tags, estimated_effort, target_release
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`, req.Name, req.Description, req.CustomerName, req.CustomersNumberOfRequests,
		req.JiraLink, req.JiraKey, req.Status, req.Priority, req.CreatedBy,
		req.Tags, req.EstimatedEffort, req.TargetRelease).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to create feature request: %w", err)
	}

	log.Printf("✅ Created feature request: id=%s name=%s customer=%s", id, req.Name, req.CustomerName)
	return id, nil
}

// GetAllFeatureRequests retrieves all feature requests from the database
func GetAllFeatureRequests() ([]FeatureRequestDB, error) {
	rows, err := DB.Query(`
		SELECT
			id, name, description, customer_name, customers_number_of_requests,
			jira_link, jira_key, status, priority,
			created_at, updated_at, created_by, tags, estimated_effort, target_release
		FROM feature_requests
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query feature requests: %w", err)
	}
	defer rows.Close()

	var requests []FeatureRequestDB
	for rows.Next() {
		var req FeatureRequestDB
		err := rows.Scan(
			&req.ID, &req.Name, &req.Description, &req.CustomerName, &req.CustomersNumberOfRequests,
			&req.JiraLink, &req.JiraKey, &req.Status, &req.Priority,
			&req.CreatedAt, &req.UpdatedAt, &req.CreatedBy, &req.Tags, &req.EstimatedEffort, &req.TargetRelease,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feature request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetPendingFeatureRequests retrieves feature requests with status 'pending'
func GetPendingFeatureRequests() ([]FeatureRequestDB, error) {
	rows, err := DB.Query(`
		SELECT
			id, name, description, customer_name, customers_number_of_requests,
			jira_link, jira_key, status, priority,
			created_at, updated_at, created_by, tags, estimated_effort, target_release
		FROM feature_requests
		WHERE status = 'pending'
		ORDER BY customers_number_of_requests DESC, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending feature requests: %w", err)
	}
	defer rows.Close()

	var requests []FeatureRequestDB
	for rows.Next() {
		var req FeatureRequestDB
		err := rows.Scan(
			&req.ID, &req.Name, &req.Description, &req.CustomerName, &req.CustomersNumberOfRequests,
			&req.JiraLink, &req.JiraKey, &req.Status, &req.Priority,
			&req.CreatedAt, &req.UpdatedAt, &req.CreatedBy, &req.Tags, &req.EstimatedEffort, &req.TargetRelease,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feature request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// UpdateFeatureRequest updates an existing feature request
func UpdateFeatureRequest(id string, req FeatureRequestDB) error {
	_, err := DB.Exec(`
		UPDATE feature_requests SET
			name = $1,
			description = $2,
			customer_name = $3,
			customers_number_of_requests = $4,
			jira_link = $5,
			jira_key = $6,
			status = $7,
			priority = $8,
			tags = $9,
			estimated_effort = $10,
			target_release = $11,
			updated_at = NOW()
		WHERE id = $12
	`, req.Name, req.Description, req.CustomerName, req.CustomersNumberOfRequests,
		req.JiraLink, req.JiraKey, req.Status, req.Priority,
		req.Tags, req.EstimatedEffort, req.TargetRelease, id)

	if err != nil {
		return fmt.Errorf("failed to update feature request: %w", err)
	}

	log.Printf("✅ Updated feature request: id=%s", id)
	return nil
}
