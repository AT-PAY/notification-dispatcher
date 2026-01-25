package api

import (
	"encoding/json"
	"log"
	"net/http"
	"notification-dispatcher/internal/config"
	"notification-dispatcher/internal/dispatcher"
	"notification-dispatcher/internal/models"
	"notification-dispatcher/internal/service"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handle struct {
	Dispatcher *dispatcher.Dispatcher
	Config     *config.Config
	svc        *service.NotificationService
}

func NewHandle(d *dispatcher.Dispatcher, cfg *config.Config, svc *service.NotificationService) *Handle {
	return &Handle{
		Dispatcher: d,
		Config:     cfg,
		svc:        svc,
	}
}

func (h *Handle) SendNotificationHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert request to Notification model
	payloadBytes, err := json.Marshal(req.Data)
	if err != nil {
		http.Error(w, "Failed to marshal payload", http.StatusBadRequest)
		return
	}

	notification := models.Notification{
		UserID:        req.UserID,
		EventType:     req.EventType,
		Payload:       payloadBytes,
		CorrelationID: req.CorrelationID,
		Channels:      req.Channels,
		CreatedAt:     time.Now(),
	}

	// Process immediately (save deliveries)
	masterID, err := h.svc.ProcessNotification(r.Context(), notification)
	if err != nil {
		http.Error(w, "Failed to process notification", http.StatusInternalServerError)
		return
	}

	// Update notification with master ID for dispatcher
	notification.ID = masterID

	// Publish to Redis for distribution across nodes
	errPR := h.Dispatcher.PublishToRedis(r.Context(), notification)
	if errPR != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, err = w.Write([]byte(`{"status": "accepted", "message": "Notification queued for dispatch"}`))
	if err != nil {
		return
	}
}

func (h *Handle) WSHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}

	client := &dispatcher.Client{
		UserID:   userID,
		SendChan: make(chan models.NotificationResponse, 256),
	}

	h.Dispatcher.Registry.Register(client.UserID, client)

	defer func() {
		h.Dispatcher.Registry.Unregister(userID, client)
		err := conn.Close()
		if err != nil {
			return
		}
		log.Printf("User %s disconnected", userID)
	}()

	log.Printf("User %s connected via WebSocket", userID)

	for msg := range client.SendChan {
		jsonData, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Printf("Error writing to websocket for user %s: %v", userID, err)
			break
		}
	}
}
