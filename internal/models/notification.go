package models

type NotificationMessage struct {
	UserId     string `json:"user_id"`
	Message    string `json:"message"`
	EventType  string `json:"event_type"`
	Timestamp  *int64 `json:"timestamp"`
	Priority   *int64 `json:"priority"`
	TimeToLive *int64 `json:"time_to_live"`
}
