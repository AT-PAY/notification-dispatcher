package dispatcher

import (
	"log"
	"notification-dispatcher/internal/models"
	"sync"
)

type Dispatcher struct {
	IngestionChan chan models.NotificationMessage
	Registry      *Registry
}

type Client struct {
	UserID   string
	SendChan chan models.NotificationMessage
}

type Registry struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewDispatcher(capacity int) *Dispatcher {
	return &Dispatcher{
		IngestionChan: make(chan models.NotificationMessage, capacity),
		Registry:      newRegistry(),
	}
}

func (d *Dispatcher) StartWorkerPool(numberWorkers int) {
	log.Printf("Starting %d dispatch workers...", numberWorkers)

	for i := 0; i < numberWorkers; i++ {
		go d.worker(i)
	}
}

func (d *Dispatcher) worker(id int) {
	for msg := range d.IngestionChan {
		log.Printf("[Worker %d] Received message for User: %s, Type: %s", id, msg.UserId, msg.EventType)

		switch msg.EventType {
		case "WEB_SOCKET":
			client, ok := d.Registry.GetClient(msg.UserId)
			if ok {
				client.SendChan <- msg
			} else {
				log.Printf("[Worker %d] User %s is offline. Message dropped.", id, msg.UserId)
			}
		default:
			log.Printf("COMING IN THE FUTURE...")
		}
	}
}

func newRegistry() *Registry {
	return &Registry{clients: make(map[string]*Client)}
}

func (r *Registry) Register(userID string, client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[userID] = client
}

func (r *Registry) Unregister(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, userID)
}

func (r *Registry) GetClient(userID string) (*Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, ok := r.clients[userID]
	return client, ok
}
