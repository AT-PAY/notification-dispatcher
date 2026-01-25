package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"notification-dispatcher/internal/dispatcher"
	"notification-dispatcher/internal/models"
	"notification-dispatcher/internal/service"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	disp   *dispatcher.Dispatcher
	svc    *service.NotificationService
}

func NewKafkaConsumer(brokers []string, topic, groupID string, d *dispatcher.Dispatcher, svc *service.NotificationService) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
		disp: d,
		svc:  svc,
	}
}

func (kc *KafkaConsumer) Start(ctx context.Context) {
	log.Println("🚀 Kafka Consumer starting...")

	for {
		msg, err := kc.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ Kafka read error: %v", err)
			return
		}

		var notificationReq models.NotificationRequest
		if err := json.Unmarshal(msg.Value, &notificationReq); err != nil {
			log.Printf("❌ Failed to unmarshal Kafka message: %v", err)
			continue
		}

		log.Printf("📨 Received Kafka message for user: %s, EventType: %s", notificationReq.UserID, notificationReq.EventType)

		// Convert NotificationRequest to Notification model
		payloadBytes, err := json.Marshal(notificationReq.Data)
		if err != nil {
			log.Printf("❌ Failed to marshal payload: %v", err)
			continue
		}

		notification := models.Notification{
			UserID:        notificationReq.UserID,
			EventType:     notificationReq.EventType,
			Payload:       payloadBytes,
			CorrelationID: notificationReq.CorrelationID,
		}

		// Process notification (save to DB and create deliveries)
		masterID, err := kc.svc.ProcessNotification(ctx, notification)
		if err != nil {
			log.Printf("❌ Failed to process notification: %v", err)
			continue
		}

		// Update notification with master ID for dispatcher
		notification.ID = masterID

		// Publish to Redis for other nodes
		if err := kc.disp.PublishToRedis(ctx, notification); err != nil {
			log.Printf("❌ Failed to publish to Redis: %v", err)
			continue
		}

		log.Printf("✅ Notification processed and published: ID=%s, UserID=%s", masterID, notification.UserID)
	}
}

func (kc *KafkaConsumer) Close() error {
	if err := kc.reader.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka reader: %w", err)
	}
	log.Println("✅ Kafka Consumer closed")
	return nil
}
