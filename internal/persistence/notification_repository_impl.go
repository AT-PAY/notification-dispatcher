package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"notification-dispatcher/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type notificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository creates and returns a new implementation of NotificationDB
// This follows the Repository pattern similar to Spring Data JPA
func NewNotificationRepository(db *sql.DB) NotificationDB {
	return &notificationRepository{
		db: sqlx.NewDb(db, "postgres"),
	}
}

// GetTemplatesByEvent retrieves all templates for a specific event type
func (r *notificationRepository) GetTemplatesByEvent(ctx context.Context, eventType models.EventType) ([]models.NotificationTemplate, error) {
	query := `
		SELECT id, event_type, channel, language, title_template, content_template, created_at
		FROM notification_templates
		WHERE event_type = $1
	`

	var templates []models.NotificationTemplate
	err := r.db.SelectContext(ctx, &templates, query, string(eventType))
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get templates for event %s: %w", eventType, err)
	}

	log.Printf("📋 Retrieved %d templates for EventType=%s", len(templates), eventType)
	return templates, nil
}

// SaveMasterNotification persists the original notification request
func (r *notificationRepository) SaveMasterNotification(ctx context.Context, noti *models.Notification) (string, error) {
	query := `
  INSERT INTO notifications (user_id, event_type, payload, correlation_id, channels, created_at)
  VALUES ($1, $2, $3, $4, $5, $6)
  RETURNING id
 `

	var id string
	err := r.db.QueryRowContext(
		ctx,
		query,
		noti.UserID,
		noti.EventType,
		noti.Payload,
		noti.CorrelationID,
		pq.Array(noti.Channels),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to save master notification: %w", err)
	}

	log.Printf("✅ Master notification saved: ID=%s, UserID=%s, Channels=%v", id, noti.UserID, noti.Channels)
	return id, nil
}

// CreateDeliveries performs batch insert of delivery records
func (r *notificationRepository) CreateDeliveries(ctx context.Context, deliveries []models.NotificationDelivery) error {
	if len(deliveries) == 0 {
		return nil
	}

	query := `
		INSERT INTO notification_deliveries (notification_id, channel, title, content, status, created_at)
		VALUES (:notification_id, :channel, :title, :content, :status, :created_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, deliveries)
	if err != nil {
		return fmt.Errorf("failed to batch insert deliveries: %w", err)
	}

	log.Printf("📤 Created %d delivery records", len(deliveries))
	return nil
}

// UpdateDeliveryStatus updates the status of a specific delivery
func (r *notificationRepository) UpdateDeliveryStatus(ctx context.Context, deliveryID string, status models.DeliveryStatus, errorDetail *string) error {
	query := `
		UPDATE notification_deliveries
		SET status = $1,    error_detail = $2,
		    sent_at = CASE WHEN $1 = 'SENT' THEN NOW() ELSE sent_at END,
		    updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, string(status), errorDetail, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found: %s", deliveryID)
	}

	log.Printf("🔄 Delivery status updated: ID=%s, Status=%s", deliveryID, status)
	return nil
}

func (r *notificationRepository) GetDeliveriesByNotificationID(ctx context.Context, notificationID string) ([]models.NotificationDelivery, error) {
	query := `
        SELECT id, notification_id, channel, title, content, status, error_detail, sent_at, read_at, created_at
        FROM notification_deliveries
        WHERE notification_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, notificationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []models.NotificationDelivery
	for rows.Next() {
		var d models.NotificationDelivery
		if err := rows.Scan(&d.ID, &d.NotificationID, &d.Channel, &d.Title, &d.Content, &d.Status, &d.ErrorDetail, &d.SentAt, &d.ReadAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

// MarkAsRead records when a user reads a notification
func (r *notificationRepository) MarkAsRead(ctx context.Context, deliveryID string) error {
	query := `
		UPDATE notification_deliveries
		SET status = $1, read_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, string(models.StatusRead), deliveryID)
	if err != nil {
		return fmt.Errorf("failed to mark delivery as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found: %s", deliveryID)
	}

	log.Printf("👁️ Delivery marked as read: ID=%s", deliveryID)
	return nil
}

// UpsertRetryRecord creates or updates retry tracking for failed deliveries
func (r *notificationRepository) UpsertRetryRecord(ctx context.Context, retry *models.NotificationRetry) error {
	query := `
		INSERT INTO notification_retries (delivery_id, retry_count, max_retries, last_error, next_retry_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (delivery_id) DO UPDATE SET
			retry_count = $2,
			last_error = $4,
			next_retry_at = $5,
			status = $6,
			updated_at = $8
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		retry.DeliveryID,
		retry.RetryCount,
		retry.MaxRetries,
		retry.LastError,
		retry.NextRetryAt,
		retry.Status,
		retry.CreatedAt,
		retry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert retry record: %w", err)
	}

	log.Printf("🔁 Retry record upserted: DeliveryID=%s, RetryCount=%d, NextRetry=%v", retry.DeliveryID, retry.RetryCount, retry.NextRetryAt)
	return nil
}

// GetPendingRetries fetches retries eligible for retry based on time and count
func (r *notificationRepository) GetPendingRetries(ctx context.Context, limit int) ([]models.NotificationRetry, error) {
	query := `
		SELECT id, delivery_id, retry_count, max_retries, last_error, next_retry_at, status, created_at, updated_at
		FROM notification_retries
		WHERE status = $1 AND next_retry_at <= NOW() AND retry_count < max_retries
		ORDER BY next_retry_at ASC
		LIMIT $2
	`

	var retries []models.NotificationRetry
	err := r.db.SelectContext(ctx, &retries, query, string(models.StatusPending), limit)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get pending retries: %w", err)
	}

	log.Printf("📥 Found %d pending retries", len(retries))
	return retries, nil
}

// GetHistoryByUserID retrieves paginated notification delivery history for a user
func (r *notificationRepository) GetHistoryByUserID(ctx context.Context, userID string, limit, offset int) ([]models.NotificationDelivery, error) {
	query := `
		SELECT id, notification_id, channel, title, content, status, error_detail, sent_at, read_at, created_at
		FROM notification_deliveries
		WHERE notification_id IN (
			SELECT id FROM notifications WHERE user_id = $1
		)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var deliveries []models.NotificationDelivery
	err := r.db.SelectContext(ctx, &deliveries, query, userID, limit, offset)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get notification history for user=%s: %w", userID, err)
	}

	log.Printf("📚 Retrieved %d notifications for user=%s", len(deliveries), userID)
	return deliveries, nil
}
