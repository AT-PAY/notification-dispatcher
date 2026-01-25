package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"notification-dispatcher/internal/models"
	"notification-dispatcher/internal/persistence"
	"slices"
	"time"
)

type NotificationService struct {
	db       persistence.NotificationDB
	renderer Renderer
}

func NewNotificationService(db persistence.NotificationDB, renderer Renderer) *NotificationService {
	return &NotificationService{
		db:       db,
		renderer: renderer,
	}
}

// ProcessNotification handles the complete notification workflow
// 1. Save master notification
// 2. Get templates for all channels
// 3. Render content with dynamic data
// 4. Create delivery records
// 5. Return master notification ID
func (ns *NotificationService) ProcessNotification(ctx context.Context, req models.Notification) (string, error) {
	// Step 1: Save master notification
	masterID, err := ns.db.SaveMasterNotification(ctx, &req)
	if err != nil {
		return "", fmt.Errorf("failed to save master notification: %w", err)
	}

	log.Printf("✅ Master notification saved: ID=%s, UserID=%s, EventType=%s", masterID, req.UserID, req.EventType)

	// Step 2: Get templates for this event type
	templates, err := ns.db.GetTemplatesByEvent(ctx, models.EventType(req.EventType))
	if err != nil {
		return "", fmt.Errorf("failed to get templates for event %s: %w", req.EventType, err)
	}

	if len(templates) == 0 {
		log.Printf("⚠️ No templates found for EventType=%s", req.EventType)
		return masterID, nil
	}

	// Step 3: Parse payload for template rendering
	var payloadData map[string]interface{}
	if err := json.Unmarshal(req.Payload, &payloadData); err != nil {
		return "", fmt.Errorf("failed to parse notification payload: %w", err)
	}

	// Step 4: Create delivery records for each channel
	deliveries := make([]models.NotificationDelivery, 0, len(templates))

	for _, template := range templates {
		if !slices.Contains(req.Channels, template.Channel) {
			log.Printf("ℹ️ Skipping channel %s as it's not in requested channels", template.Channel)
			continue
		}

		// Render title and content with dynamic data
		title := ns.renderer.Render(template.TitleTemplate, payloadData)
		content := ns.renderer.Render(template.ContentTemplate, payloadData)

		delivery := models.NotificationDelivery{
			NotificationID: masterID,
			Channel:        template.Channel,
			Title:          title,
			Content:        content,
			Status:         string(models.StatusPending),
			CreatedAt:      time.Now(),
		}

		deliveries = append(deliveries, delivery)
		log.Printf("📋 Prepared delivery: Channel=%s, Title=%s", template.Channel, title)
	}

	// Batch insert all deliveries
	if err := ns.db.CreateDeliveries(ctx, deliveries); err != nil {
		return "", fmt.Errorf("failed to create deliveries: %w", err)
	}

	log.Printf("📤 Created %d delivery records for notification=%s", len(deliveries), masterID)
	return masterID, nil
}

// UpdateDeliverySuccess marks a delivery as successfully sent
func (ns *NotificationService) UpdateDeliverySuccess(ctx context.Context, deliveryID string) error {
	return ns.db.UpdateDeliveryStatus(ctx, deliveryID, models.DeliveryStatus(models.StatusSent), nil)
}

// UpdateDeliveryFailure marks a delivery as failed and logs the error
func (ns *NotificationService) UpdateDeliveryFailure(ctx context.Context, deliveryID string, errorDetail string) error {
	return ns.db.UpdateDeliveryStatus(ctx, deliveryID, models.DeliveryStatus(models.StatusFailed), &errorDetail)
}

func (ns *NotificationService) GetNotificationDeliveries(ctx context.Context, notificationID string) ([]models.NotificationDelivery, error) {
	deliveries, err := ns.db.GetDeliveriesByNotificationID(ctx, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries for notification=%s: %w", notificationID, err)
	}

	log.Printf("📥 Retrieved %d deliveries for notification=%s", len(deliveries), notificationID)
	return deliveries, nil
}

// HandleRetry manages the retry logic for failed deliveries
// Uses exponential backoff strategy
func (ns *NotificationService) HandleRetry(ctx context.Context, deliveryID string, retryCount int, maxRetries int, lastError string, backoffSeconds int) error {
	retry := &models.NotificationRetry{
		DeliveryID:  deliveryID,
		RetryCount:  retryCount,
		MaxRetries:  maxRetries,
		LastError:   &lastError,
		NextRetryAt: time.Now().Add(time.Duration(backoffSeconds) * time.Second),
		Status:      string(models.StatusPending),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := ns.db.UpsertRetryRecord(ctx, retry); err != nil {
		return fmt.Errorf("failed to upsert retry record: %w", err)
	}

	log.Printf("🔁 Retry scheduled: DeliveryID=%s, NextRetry=%v, Backoff=%ds", deliveryID, retry.NextRetryAt, backoffSeconds)
	return nil
}

// ProcessPendingRetries fetches and processes failed deliveries eligible for retry
func (ns *NotificationService) ProcessPendingRetries(ctx context.Context, limit int) ([]models.NotificationRetry, error) {
	retries, err := ns.db.GetPendingRetries(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending retries: %w", err)
	}

	log.Printf("📥 Found %d pending retries", len(retries))
	return retries, nil
}

// GetNotificationHistory retrieves notification history for a user
func (ns *NotificationService) GetNotificationHistory(ctx context.Context, userID string, limit, offset int) ([]models.NotificationDelivery, error) {
	deliveries, err := ns.db.GetHistoryByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notification history for user=%s: %w", userID, err)
	}

	log.Printf("📚 Retrieved %d notifications for user=%s", len(deliveries), userID)
	return deliveries, nil
}

// MarkAsRead updates a delivery as read by the user
func (ns *NotificationService) MarkAsRead(ctx context.Context, deliveryID string) error {
	if err := ns.db.MarkAsRead(ctx, deliveryID); err != nil {
		return fmt.Errorf("failed to mark delivery as read: %w", err)
	}

	log.Printf("👁️ Delivery marked as read: ID=%s", deliveryID)
	return nil
}
