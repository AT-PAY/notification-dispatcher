package models

import (
	"encoding/json"
	"time"
)

type NotificationTemplate struct {
	ID              string    `db:"id"`
	EventType       string    `db:"event_type"`
	Channel         string    `db:"channel"`
	Language        string    `db:"language"`
	TitleTemplate   string    `db:"title_template"`
	ContentTemplate string    `db:"content_template"`
	CreatedAt       time.Time `db:"created_at"`
	CreatedBy       string    `db:"created_by"`
	UpdatedAt       time.Time `db:"updated_at"`
	UpdatedBy       string    `db:"updated_by"`
}

type Notification struct {
	ID            string          `db:"id"`
	UserID        string          `db:"user_id"`
	EventType     string          `db:"event_type"`
	Payload       json.RawMessage `db:"payload"`
	CorrelationID string          `db:"correlation_id"`
	Channels      []string        `json:"channels"`
	CreatedAt     time.Time       `db:"created_at"`
	CreatedBy     string          `db:"created_by"`
	UpdatedAt     time.Time       `db:"updated_at"`
	UpdatedBy     string          `db:"updated_by"`
}

type NotificationDelivery struct {
	ID             string `db:"id"`
	NotificationID string `db:"notification_id"`
	Channel        string `db:"channel"`

	TitleVi   string `db:"title_vi"`
	TitleEn   string `db:"title_en"`
	ContentVi string `db:"content_vi"`
	ContentEn string `db:"content_en"`

	Status      string     `db:"status"`
	ErrorDetail *string    `db:"error_detail"`
	SentAt      *time.Time `db:"sent_at"`
	ReadAt      *time.Time `db:"read_at"`
	CreatedAt   time.Time  `db:"created_at"`
	CreatedBy   string     `db:"created_by"`
	UpdatedAt   time.Time  `db:"updated_at"`
	UpdatedBy   string     `db:"updated_by"`
}

type NotificationRetry struct {
	ID          string    `db:"id"`
	DeliveryID  string    `db:"delivery_id"`
	RetryCount  int       `db:"retry_count"`
	MaxRetries  int       `db:"max_retries"`
	LastError   *string   `db:"last_error"`
	NextRetryAt time.Time `db:"next_retry_at"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	CreatedBy   string    `db:"created_by"`
	UpdatedAt   time.Time `db:"updated_at"`
	UpdatedBy   string    `db:"updated_by"`
}
