package models

import (
	"encoding/json"
	"time"
)

type NotificationRequest struct {
	UserID        string                 `json:"user_id"`
	EventType     string                 `json:"event_type"`
	Data          map[string]interface{} `json:"data"`
	CorrelationID string                 `json:"correlation_id"`
	Channels      []string               `json:"channels"`
}

type NotificationResponse struct {
	ID                   string
	UserID               string
	EventType            string
	Payload              json.RawMessage
	CorrelationID        string
	Channels             []string
	CreatedAt            time.Time
	CreatedBy            string
	UpdatedAt            time.Time
	UpdatedBy            string
	NotificationDelivery NotificationDelivery
}
