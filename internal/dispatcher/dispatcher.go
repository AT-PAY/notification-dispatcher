package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"notification-dispatcher/internal/models"
	"notification-dispatcher/internal/service"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Dispatcher struct {
	IngestionChan       chan models.NotificationResponse
	Registry            *Registry
	WG                  sync.WaitGroup
	RedisClient         *redis.Client
	RedisChannel        string
	NotificationService *service.NotificationService
}

type Client struct {
	UserID   string
	SendChan chan models.NotificationResponse
}

type Registry struct {
	clients map[string]map[*Client]bool
	mu      sync.RWMutex
}

func NewDispatcher(capacity int, redisAddr string, svc *service.NotificationService) *Dispatcher {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &Dispatcher{
		IngestionChan:       make(chan models.NotificationResponse, capacity),
		Registry:            newRegistry(),
		RedisClient:         rdb,
		RedisChannel:        "notifications",
		NotificationService: svc,
	}
}

func (d *Dispatcher) StartRedisSubscriber() {
	ctx := context.Background()
	pubsub := d.RedisClient.Subscribe(ctx, d.RedisChannel)

	log.Printf("✅ Subscribed to Redis channel: %s", d.RedisChannel)

	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var notification models.Notification
			if err := json.Unmarshal([]byte(msg.Payload), &notification); err != nil {
				log.Printf("❌ Error unmarshaling Redis message: %v", err)
				continue
			}

			deliveries, err := d.NotificationService.GetNotificationDeliveries(ctx, notification.ID)
			if err != nil {
				log.Printf("❌ Error getting deliveries for notification %s: %v", notification.ID, err)
				continue
			}

			var notificationResponse models.NotificationResponse
			notificationResponse.ID = notification.ID
			notificationResponse.UserID = notification.UserID
			notificationResponse.EventType = notification.EventType
			notificationResponse.Payload = notification.Payload
			notificationResponse.CorrelationID = notification.CorrelationID
			notificationResponse.Channels = notification.Channels
			notificationResponse.CreatedAt = notification.CreatedAt
			notificationResponse.CreatedBy = notification.CreatedBy
			notificationResponse.UpdatedAt = notification.UpdatedAt
			notificationResponse.UpdatedBy = notification.UpdatedBy

			for _, delivery := range deliveries {
				notificationResponse.NotificationDelivery = delivery
				d.IngestionChan <- notificationResponse
			}

		}
	}()
}

func (d *Dispatcher) PublishToRedis(ctx context.Context, msg models.Notification) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	err = d.RedisClient.Publish(ctx, d.RedisChannel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to Redis: %w", err)
	}

	log.Printf("📤 Published to Redis: UserID=%s, EventType=%s", msg.UserID, msg.EventType)
	return nil
}

func (d *Dispatcher) StartWorkerPool(numberWorkers int) {
	log.Printf("🚀 Starting %d dispatch workers...", numberWorkers)

	for i := 0; i < numberWorkers; i++ {
		d.WG.Add(1)
		go d.worker(i)
	}
}

func (d *Dispatcher) worker(id int) {
	defer d.WG.Done()

	for msg := range d.IngestionChan {
		log.Printf("[Worker %d] 🔄 Processing: UserID=%s, EventType=%s", id, msg.UserID, msg.EventType)

		// Try to find connected clients for this user
		clients := d.Registry.GetClients(msg.UserID)

		if msg.NotificationDelivery.Channel == "WEB_SOCKET" {
			if len(clients) > 0 {
				log.Printf("[Worker %d] 📍 Registry: UserID=%s has %d connection(s)", id, msg.UserID, len(clients))
				// User online: send immediately
				for _, client := range clients {
					select {
					case client.SendChan <- models.NotificationResponse{}:
						log.Printf("[Worker %d] ✅ Sent to WebSocket for UserID=%s", id, msg.UserID)
					case <-time.After(2 * time.Second):
						log.Printf("[Worker %d] ⚠️ Send channel timeout for UserID=%s", id, msg.UserID)
					}
				}
			} else {
				// User offline: Notification đã được lưu bởi Kafka/API handler
				log.Printf("[Worker %d] ℹ️ User %s not online, message stored in database", id, msg.UserID)
			}
		}

		// Process other channels (Email, SMS, Push)
		if d.isChannelEnabled(msg, "EMAIL") {
			d.handleEmailChannel(id, msg)
		}
		if d.isChannelEnabled(msg, "SMS") {
			d.handleSMSChannel(id, msg)
		}
		if d.isChannelEnabled(msg, "PUSH") {
			d.handlePushChannel(id, msg)
		}
	}

	log.Printf("[Worker %d] ✔️ Cleaned up and exited", id)
}

// isChannelEnabled checks if a specific channel is enabled for this notification
func (d *Dispatcher) isChannelEnabled(msg models.NotificationResponse, channel string) bool {
	if len(msg.Channels) == 0 {
		return channel == "WEB_SOCKET" // Default to WebSocket if no channels specified
	}
	for _, ch := range msg.Channels {
		if ch == channel {
			return true
		}
	}
	return false
}

// handleEmailChannel sends notification via email
func (d *Dispatcher) handleEmailChannel(workerID int, msg models.NotificationResponse) {
	log.Printf("[Worker %d] 📧 Sending EMAIL for UserID=%s", workerID, msg.UserID)
	// TODO: Implement email sending logic
	// - Get email address from UserID
	// - Call email service/API
	// - Update delivery status in DB
}

// handleSMSChannel sends notification via SMS
func (d *Dispatcher) handleSMSChannel(workerID int, msg models.NotificationResponse) {
	log.Printf("[Worker %d] 📱 Sending SMS for UserID=%s", workerID, msg.UserID)
	// TODO: Implement SMS sending logic
	// - Get phone number from UserID
	// - Call SMS service/API
	// - Update delivery status in DB
}

// handlePushChannel sends notification via push notification
func (d *Dispatcher) handlePushChannel(workerID int, msg models.NotificationResponse) {
	log.Printf("[Worker %d] 🔔 Sending PUSH for UserID=%s", workerID, msg.UserID)
	// TODO: Implement push notification logic
	// - Get device tokens from UserID
	// - Call FCM/APNs service
	// - Update delivery status in DB
}

func (d *Dispatcher) Shutdown() {
	log.Println("🛑 Shutting down dispatcher workers...")
	close(d.IngestionChan)
	d.WG.Wait()
	log.Println("✅ All workers finished.")
}

func newRegistry() *Registry {
	return &Registry{clients: make(map[string]map[*Client]bool)}
}

func (r *Registry) Register(userID string, client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[userID]; !ok {
		r.clients[userID] = make(map[*Client]bool)
	}
	r.clients[userID][client] = true
	log.Printf("📍 Registry: UserID=%s registered. Total connections: %d", userID, len(r.clients[userID]))
}

func (r *Registry) Unregister(userID string, client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if clients, ok := r.clients[userID]; ok {
		delete(clients, client)

		if len(clients) == 0 {
			delete(r.clients, userID)
			log.Printf("📍 Registry: UserID=%s unregistered (no more connections)", userID)
		}
	}
}

func (r *Registry) GetClients(userID string) []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clientsMap, ok := r.clients[userID]
	if !ok {
		return nil
	}

	clients := make([]*Client, 0, len(clientsMap))
	for client := range clientsMap {
		clients = append(clients, client)
	}
	return clients
}
