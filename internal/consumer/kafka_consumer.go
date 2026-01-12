package consumer

import (
	"context"
	"encoding/json"
	"log"
	"notification-dispatcher/internal/dispatcher"
	"notification-dispatcher/internal/models"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	disp   *dispatcher.Dispatcher
}

func NewKafkaConsumer(brokers []string, topic, groupID string, d *dispatcher.Dispatcher) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
		disp: d,
	}
}

func (kc *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Kafka Consumer starting...")
	for {
		m, err := kc.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			return
		}

		var req models.NotificationMessage
		if err := json.Unmarshal(m.Value, &req); err != nil {
			log.Printf("Failed to unmarshal Kafka message: %v", err)
			continue
		}

		log.Printf("Received Kafka message for user: %s", req.UserId)
		kc.disp.PublishToRedis(ctx, req)
	}
}

func (kc *KafkaConsumer) Close() error {
	return kc.reader.Close()
}
