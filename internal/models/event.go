package models

type EventType string
type DeliveryChannel string
type DeliveryStatus string

const (
	// Business Event Types
	EventPaymentSuccess     EventType = "PAYMENT_SUCCESS"
	EventBalanceFluctuation EventType = "BALANCE_FLUCTUATION"
	EventOTP                EventType = "OTP_CODE"

	// Delivery Channels
	ChannelWebSocket DeliveryChannel = "WEB_SOCKET"
	ChannelPush      DeliveryChannel = "PUSH_NOTIFICATION"
	ChannelSMS       DeliveryChannel = "SMS"
	ChannelEmail     DeliveryChannel = "EMAIL"

	// Statuses
	StatusPending   DeliveryStatus = "PENDING"
	StatusSent      DeliveryStatus = "SENT"
	StatusFailed    DeliveryStatus = "FAILED"
	StatusCompleted DeliveryStatus = "COMPLETED"
	StatusRead      DeliveryStatus = "READ"
)
