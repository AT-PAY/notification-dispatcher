package persistence

import (
	"context"
	"notification-dispatcher/internal/models"
)

// NotificationDB defines the data access layer for the notification system.
// It handles templates, master records, delivery tracking, and retry logic.
type NotificationDB interface {
	// --- Template Management ---

	// GetTemplatesByEvent retrieves all delivery configurations (channels)
	// for a specific business event (e.g., PAYMENT_SUCCESS, OTP_CODE).
	GetTemplatesByEvent(ctx context.Context, eventType models.EventType) ([]models.NotificationTemplate, error)

	// --- Notification Persistence ---

	// SaveMasterNotification persists the original notification request from the source service.
	// Returns the generated unique identifier (UUID) for the master record.
	SaveMasterNotification(ctx context.Context, noti *models.Notification) (string, error)

	// CreateDeliveries performs a batch insert of delivery records for multiple channels.
	// Each record represents a specific attempt to reach the user via a specific medium.
	CreateDeliveries(ctx context.Context, deliveries []models.NotificationDelivery) error

	// UpdateDeliveryStatus updates the state of a specific delivery channel.
	// Used to transition records to SENT, FAILED, or DELIVERED status.
	UpdateDeliveryStatus(ctx context.Context, deliveryID string, status models.DeliveryStatus, errorDetail *string) error

	GetDeliveriesByNotificationID(ctx context.Context, notificationID string) ([]models.NotificationDelivery, error)

	// MarkAsRead records the timestamp when a user interacts with/opens a specific notification.
	MarkAsRead(ctx context.Context, deliveryID string) error

	// --- Reliability & Retry Logic ---

	// UpsertRetryRecord creates or updates a retry tracking entry for a failed delivery.
	// Used to manage the exponential backoff strategy and retry counts.
	UpsertRetryRecord(ctx context.Context, retry *models.NotificationRetry) error

	// GetPendingRetries fetches failed delivery records that are eligible for a retry attempt
	// based on the current time and their scheduled next_retry_at.
	GetPendingRetries(ctx context.Context, limit int) ([]models.NotificationRetry, error)

	// --- Audit & History ---

	// GetHistoryByUserID retrieves a paginated list of successful notification deliveries for a user.
	// Typically used to populate the "Inbox" or "Notification Center" in the client application.
	GetHistoryByUserID(ctx context.Context, userID string, limit, offset int) ([]models.NotificationDelivery, error)
}
