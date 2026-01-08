package dispatcher

import (
	"log"
	"notification-dispatcher/internal/models"
	"sync"
)

type Dispatcher struct {
	IngestionChan chan models.NotificationMessage
	Registry      *Registry
	WG            sync.WaitGroup
}

type Client struct {
	UserID   string
	SendChan chan models.NotificationMessage
}

type Registry struct {
	clients map[string]map[*Client]bool
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
		d.WG.Add(1)
		go d.worker(i)
	}
}

func (d *Dispatcher) worker(id int) {
	defer d.WG.Done()

	for msg := range d.IngestionChan {
		log.Printf("[Worker %d] Processing: User=%s, Type=%s", id, msg.UserId, msg.EventType)

		switch msg.EventType {
		case "WEB_SOCKET":
			clients := d.Registry.GetClients(msg.UserId)

			if clients != nil {
				for _, client := range clients {
					select {
					case client.SendChan <- msg:
						log.Printf("[Worker %d] ✅ Sent to device for %s", id, msg.UserId)
					default:
						log.Printf("[Worker %d] ⚠️ Buffer full for device of %s", id, msg.UserId)
					}
				}
			} else {
				log.Printf("[Worker %d] ❌ User %s not online", id, msg.UserId)
			}
		default:
			log.Printf("[Worker %d] ⚠️ Unhandled EventType: %s", id, msg.EventType)
		}
	}
	log.Printf("[Worker %d] Cleaned up and exited", id)
}

func (d *Dispatcher) Shutdown() {
	log.Println("Shutting down dispatcher workers...")
	close(d.IngestionChan)
	d.WG.Wait()
	log.Println("All workers finished.")
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
	log.Printf("Registry: User %s added a connection. Total connections: %d", userID, len(r.clients[userID]))
}

func (r *Registry) Unregister(userID string, client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if clients, ok := r.clients[userID]; ok {
		delete(clients, client)

		if len(clients) == 0 {
			delete(r.clients, userID)
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
