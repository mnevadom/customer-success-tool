package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// ThenaRequestDB represents a Thena request in the database
type ThenaRequestDB struct {
	ID                     string
	ThenaID                string
	RequestID              int
	EventID                string
	Status                 string
	SubStatus              sql.NullString
	SubStatusName          sql.NullString
	SubStatusDescription   sql.NullString
	CustomerName           string
	CRMAccountID           sql.NullString
	CRMAccountName         sql.NullString
	ChannelID              sql.NullString
	ChannelName            sql.NullString
	Description            sql.NullString
	Permalink              sql.NullString
	ThenaURL               sql.NullString
	RequestorID            sql.NullString
	RequestorName          sql.NullString
	RequestorEmail         sql.NullString
	RequestorDomain        sql.NullString
	AssignedToID           sql.NullString
	AssignedToName         sql.NullString
	AssignedToEmail        sql.NullString
	AssignedToDomain       sql.NullString
	AssignedByID           sql.NullString
	AssignedByName         sql.NullString
	AssignedByEmail        sql.NullString
	AssignedByDomain       sql.NullString
	SenderID               sql.NullString
	SenderName             sql.NullString
	SenderEmail            sql.NullString
	SenderDomain           sql.NullString
	ReplyCount             int
	FirstResponseAt        sql.NullString
	LastCustomerMessageAt  sql.NullString
	LastVendorMessageAt    sql.NullString
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ReceivedAt             time.Time
}

// UpsertThenaRequest inserts or updates a Thena request and tracks status history
func UpsertThenaRequest(req ThenaRequestDB) error {
	// Start a transaction
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if request exists and get current status
	var existingStatus, existingSubStatus, existingSubStatusName sql.NullString
	var exists bool
	err = tx.QueryRow(`
		SELECT status, sub_status, sub_status_name
		FROM thena_requests
		WHERE thena_id = $1
	`, req.ThenaID).Scan(&existingStatus, &existingSubStatus, &existingSubStatusName)

	if err == sql.ErrNoRows {
		// New request - insert
		exists = false
	} else if err != nil {
		return fmt.Errorf("failed to check existing request: %w", err)
	} else {
		exists = true
	}

	// Upsert the request
	_, err = tx.Exec(`
		INSERT INTO thena_requests (
			thena_id, request_id, event_id, status, sub_status, sub_status_name, sub_status_description,
			customer_name, crm_account_id, crm_account_name, channel_id, channel_name,
			description, permalink, thena_url,
			requestor_id, requestor_name, requestor_email, requestor_domain,
			assigned_to_id, assigned_to_name, assigned_to_email, assigned_to_domain,
			assigned_by_id, assigned_by_name, assigned_by_email, assigned_by_domain,
			sender_id, sender_name, sender_email, sender_domain,
			reply_count, first_response_at, last_customer_message_at, last_vendor_message_at,
			created_at, updated_at, received_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27,
			$28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38
		)
		ON CONFLICT (thena_id) DO UPDATE SET
			event_id = EXCLUDED.event_id,
			status = EXCLUDED.status,
			sub_status = EXCLUDED.sub_status,
			sub_status_name = EXCLUDED.sub_status_name,
			sub_status_description = EXCLUDED.sub_status_description,
			channel_id = EXCLUDED.channel_id,
			channel_name = EXCLUDED.channel_name,
			description = EXCLUDED.description,
			permalink = EXCLUDED.permalink,
			thena_url = EXCLUDED.thena_url,
			requestor_id = EXCLUDED.requestor_id,
			requestor_name = EXCLUDED.requestor_name,
			requestor_email = EXCLUDED.requestor_email,
			requestor_domain = EXCLUDED.requestor_domain,
			assigned_to_id = EXCLUDED.assigned_to_id,
			assigned_to_name = EXCLUDED.assigned_to_name,
			assigned_to_email = EXCLUDED.assigned_to_email,
			assigned_to_domain = EXCLUDED.assigned_to_domain,
			assigned_by_id = EXCLUDED.assigned_by_id,
			assigned_by_name = EXCLUDED.assigned_by_name,
			assigned_by_email = EXCLUDED.assigned_by_email,
			assigned_by_domain = EXCLUDED.assigned_by_domain,
			sender_id = EXCLUDED.sender_id,
			sender_name = EXCLUDED.sender_name,
			sender_email = EXCLUDED.sender_email,
			sender_domain = EXCLUDED.sender_domain,
			reply_count = EXCLUDED.reply_count,
			first_response_at = EXCLUDED.first_response_at,
			last_customer_message_at = EXCLUDED.last_customer_message_at,
			last_vendor_message_at = EXCLUDED.last_vendor_message_at,
			updated_at = EXCLUDED.updated_at,
			updated_at_db = NOW()
	`, req.ThenaID, req.RequestID, req.EventID, req.Status, req.SubStatus, req.SubStatusName, req.SubStatusDescription,
		req.CustomerName, req.CRMAccountID, req.CRMAccountName, req.ChannelID, req.ChannelName,
		req.Description, req.Permalink, req.ThenaURL,
		req.RequestorID, req.RequestorName, req.RequestorEmail, req.RequestorDomain,
		req.AssignedToID, req.AssignedToName, req.AssignedToEmail, req.AssignedToDomain,
		req.AssignedByID, req.AssignedByName, req.AssignedByEmail, req.AssignedByDomain,
		req.SenderID, req.SenderName, req.SenderEmail, req.SenderDomain,
		req.ReplyCount, req.FirstResponseAt, req.LastCustomerMessageAt, req.LastVendorMessageAt,
		req.CreatedAt, req.UpdatedAt, req.ReceivedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert request: %w", err)
	}

	// Track status history if status changed or if it's a new request
	statusChanged := !exists ||
		existingStatus.String != req.Status ||
		existingSubStatus.String != req.SubStatus.String

	if statusChanged {
		var fromStatus, fromSubStatus, fromSubStatusName sql.NullString
		if exists {
			fromStatus = existingStatus
			fromSubStatus = existingSubStatus
			fromSubStatusName = existingSubStatusName
		}

		err = insertStatusHistory(tx, req, fromStatus, fromSubStatus, fromSubStatusName)
		if err != nil {
			log.Printf("⚠️  Failed to insert status history: %v", err)
			// Don't fail the whole transaction, just log the error
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if exists {
		log.Printf("✅ Updated request: thena_id=%s request_id=%d", req.ThenaID, req.RequestID)
	} else {
		log.Printf("✅ Inserted new request: thena_id=%s request_id=%d", req.ThenaID, req.RequestID)
	}

	return nil
}

// insertStatusHistory records a status change
func insertStatusHistory(tx *sql.Tx, req ThenaRequestDB, fromStatus, fromSubStatus, fromSubStatusName sql.NullString) error {
	_, err := tx.Exec(`
		INSERT INTO thena_status_history (
			thena_id, request_id, event_id,
			from_status, to_status,
			from_sub_status, to_sub_status,
			from_sub_status_name, to_sub_status_name,
			changed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (event_id, to_status, to_sub_status) DO NOTHING
	`, req.ThenaID, req.RequestID, req.EventID,
		fromStatus, req.Status,
		fromSubStatus, req.SubStatus,
		fromSubStatusName, req.SubStatusName,
		req.UpdatedAt)

	return err
}

// GetAllThenaRequests retrieves all Thena requests from the database
func GetAllThenaRequests() ([]ThenaRequestDB, error) {
	rows, err := DB.Query(`
		SELECT
			id, thena_id, request_id, event_id, status, sub_status, sub_status_name, sub_status_description,
			customer_name, crm_account_id, crm_account_name, channel_id, channel_name,
			description, permalink, thena_url,
			requestor_id, requestor_name, requestor_email, requestor_domain,
			assigned_to_id, assigned_to_name, assigned_to_email, assigned_to_domain,
			assigned_by_id, assigned_by_name, assigned_by_email, assigned_by_domain,
			sender_id, sender_name, sender_email, sender_domain,
			reply_count, first_response_at, last_customer_message_at, last_vendor_message_at,
			created_at, updated_at, received_at
		FROM thena_requests
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests: %w", err)
	}
	defer rows.Close()

	var requests []ThenaRequestDB
	for rows.Next() {
		var req ThenaRequestDB
		err := rows.Scan(
			&req.ID, &req.ThenaID, &req.RequestID, &req.EventID, &req.Status, &req.SubStatus, &req.SubStatusName, &req.SubStatusDescription,
			&req.CustomerName, &req.CRMAccountID, &req.CRMAccountName, &req.ChannelID, &req.ChannelName,
			&req.Description, &req.Permalink, &req.ThenaURL,
			&req.RequestorID, &req.RequestorName, &req.RequestorEmail, &req.RequestorDomain,
			&req.AssignedToID, &req.AssignedToName, &req.AssignedToEmail, &req.AssignedToDomain,
			&req.AssignedByID, &req.AssignedByName, &req.AssignedByEmail, &req.AssignedByDomain,
			&req.SenderID, &req.SenderName, &req.SenderEmail, &req.SenderDomain,
			&req.ReplyCount, &req.FirstResponseAt, &req.LastCustomerMessageAt, &req.LastVendorMessageAt,
			&req.CreatedAt, &req.UpdatedAt, &req.ReceivedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}
