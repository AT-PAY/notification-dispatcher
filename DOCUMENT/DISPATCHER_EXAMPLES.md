# Dispatcher & NotificationService Integration Examples

## 🔗 Integration Pattern

### **Complete Request → Response Flow**

```
1. API Handler receives POST request
   ↓
2. Dispatcher.PublishToRedis()
   ↓
3. All nodes receive via Redis Pub/Sub
   ↓
4. Worker Pool processes concurrently
   ↓
5. Check Registry for online users
   ├─ YES → Send via WebSocket immediately
   └─ NO → NotificationService saves to DB
   ↓
6. Client receives real-time notification
```

---

## 💻 Code Examples

### **Example 1: Basic Notification Flow**

```go
// cmd/main.go - Initialization

package main

import (
	"notification-dispatcher/internal/api"
	"notification-dispatcher/internal/config"
	"notification-dispatcher/internal/dispatcher"
	"net/http"
)

func main() {
	// 1. Load configuration
	cfg := config.Load()

	// 2. Create dispatcher
	disp := dispatcher.NewDispatcher(1000, "localhost:6379")

	// 3. Start Redis subscriber (in goroutine)
	disp.StartRedisSubscriber()

	// 4. Start worker pool (4 concurrent workers)
	disp.StartWorkerPool(4)

	// 5. Create API handlers
	h := api.NewHandle(disp, cfg)

	// 6. Register routes
	http.HandleFunc("/api/v1/send", h.SendNotificationHandle)
	http.HandleFunc("/ws", h.WSHandler)

	// 7. Start server
	http.ListenAndServe(":8080", nil)
}
```

---

### **Example 2: Send Notification API**

```bash
# Scenario: Payment success notification

curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "event_type": "PAYMENT_SUCCESS",
    "data": {
      "amount": 5000000,
      "currency": "VND",
      "transaction_id": "TXN-2026-001",
      "timestamp": "2026-01-18T10:30:00Z"
    },
    "correlation_id": "corr-abc-def-123"
  }'

# Response: 202 Accepted
# {
#   "status": "accepted",
#   "message": "Notification queued for dispatch"
# }

# Internal flow:
# 1. SendNotificationHandle validates request
# 2. Creates Notification struct:
#    {
#      UserID: "user-123",
#      EventType: "PAYMENT_SUCCESS",
#      Payload: {...},
#      CorrelationID: "corr-abc-def-123",
#      CreatedAt: now()
#    }
# 3. PublishToRedis() to channel "notifications"
# 4. All nodes receive via subscription
# 5. Worker checks if user online:
#    - If YES (connected via WebSocket): Send immediately
#    - If NO: Save to DB via NotificationService
```

---

### **Example 3: WebSocket Client Connection**

```javascript
// Client: Browser/Mobile App

class NotificationManager {
    constructor(userId, serverUrl = 'ws://localhost:8080') {
        this.userId = userId;
        this.serverUrl = serverUrl;
        this.connect();
    }

    connect() {
        // Connect with user ID parameter
        this.ws = new WebSocket(`${this.serverUrl}/ws?user_id=${this.userId}`);

        this.ws.onopen = () => {
            console.log(`✅ Connected as user: ${this.userId}`);
            // Dispatcher.Registry.Register() called on server side
        };

        this.ws.onmessage = (event) => {
            try {
                const notification = JSON.parse(event.data);
                console.log('📬 Received notification:', notification);
                
                this.handleNotification(notification);
            } catch (e) {
                console.error('❌ Failed to parse notification:', e);
            }
        };

        this.ws.onerror = (error) => {
            console.error('⚠️ WebSocket error:', error);
        };

        this.ws.onclose = () => {
            console.log('🔌 Disconnected');
            // Dispatcher.Registry.Unregister() called on server side
        };
    }

    handleNotification(notification) {
        const { event_type, payload } = notification;

        switch (event_type) {
            case 'PAYMENT_SUCCESS':
                this.showPaymentSuccess(payload);
                break;
            case 'OTP_CODE':
                this.showOTPCode(payload);
                break;
            case 'BALANCE_FLUCTUATION':
                this.showBalanceAlert(payload);
                break;
            default:
                console.warn('Unknown event type:', event_type);
        }
    }

    showPaymentSuccess(data) {
        // Show toast notification
        console.log(`💰 Payment successful: ${data.amount} ${data.currency}`);
        // Update UI, play sound, etc.
    }

    showOTPCode(data) {
        console.log(`🔐 Your OTP code: ${data.code}`);
    }

    showBalanceAlert(data) {
        console.log(`💳 Balance: ${data.new_balance}`);
    }
}

// Usage
const notificationMgr = new NotificationManager('user-123');
```

---

### **Example 4: Offline Storage (Phase 1)**

```go
// NotificationService handles offline users

// When user is offline (not in Registry):
// 1. Worker logs: "User not online, message stored in database"
// 2. NotificationService.ProcessNotification() is called
// 3. Flow:
//    a. SaveMasterNotification() → notifications table
//    b. GetTemplatesByEvent() → Get templates
//    c. Render() → Replace {{placeholders}} with data
//    d. CreateDeliveries() → notification_deliveries table
//    e. Return notificationID

// When user comes online:
// 1. Client calls: GET /api/v1/history?user_id=user-123&limit=20
// 2. NotificationService.GetNotificationHistory() fetches from DB
// 3. Returns paginated list of unread notifications
```

---

### **Example 5: Notification History API**

```go
// internal/api/handle.go (to be added)

// GetNotificationHistory returns unread notifications for a user
func (h *Handle) GetNotificationHistoryHandle(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("user_id")
    limit := r.URL.Query().Get("limit")    // default 20
    offset := r.URL.Query().Get("offset")  // default 0

    // Call service
    deliveries, err := h.NotificationService.GetNotificationHistory(
        r.Context(), userID, limit, offset,
    )

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data": deliveries,
    })
}

// MarkAsRead marks notification as read
func (h *Handle) MarkAsReadHandle(w http.ResponseWriter, r *http.Request) {
    deliveryID := r.URL.Query().Get("delivery_id")
    
    err := h.NotificationService.MarkAsRead(r.Context(), deliveryID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "read",
        "delivery_id": deliveryID,
    })
}
```

---

### **Example 6: Multiple Nodes Setup**

```bash
# Terminal 1: Start Node 1 (Port 8080)
go run cmd/main.go -port=8080 -node=node-1

# Terminal 2: Start Node 2 (Port 8081)
go run cmd/main.go -port=8081 -node=node-2

# Terminal 3: Client connects to Node 1
wscat -c "ws://localhost:8080/ws?user_id=user-1"

# Terminal 4: Send notification via Node 2
curl -X POST http://localhost:8081/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-1",
    "event_type": "PAYMENT_SUCCESS",
    "data": {"amount": 1000}
  }'

# Flow:
# Node 2: PublishToRedis() → Redis channel
# Node 1: Receives from Redis → IngestionChan → Worker
# Worker: Checks Registry → Finds client → Sends WebSocket
# Client: Receives notification from Node 1 ✅

# Logs:
# [Node 2] 📤 Published to Redis: UserID=user-1, EventType=PAYMENT_SUCCESS
# [Node 1] ✅ Subscribed to Redis channel: notifications
# [Node 1] [Worker 0] 🔄 Processing: UserID=user-1, EventType=PAYMENT_SUCCESS
# [Node 1] 📍 Registry: UserID=user-1 has 1 connection
# [Node 1] [Worker 0] ✅ Sent to WebSocket for UserID=user-1
```

---

### **Example 7: Logging Output**

```
# User connects to WebSocket
📍 Registry: UserID=user-1 registered. Total connections: 1
User user-1 connected via WebSocket

# Notification arrives at Node 2
✅ Subscribed to Redis channel: notifications

# Node 2 publishes notification
📤 Published to Redis: UserID=user-1, EventType=PAYMENT_SUCCESS

# Node 1 receives (subscription) and processes
🚀 Starting 4 dispatch workers...
[Worker 0] 🔄 Processing: UserID=user-1, EventType=PAYMENT_SUCCESS
[Worker 0] ✅ Sent to WebSocket for UserID=user-1

# User offline scenario
[Worker 1] 🔄 Processing: UserID=user-999, EventType=OTP_CODE
[Worker 1] ℹ️ User user-999 not online on this node, message stored in database
✅ Master notification saved: ID=notif-123, UserID=user-999, EventType=OTP_CODE
📋 Prepared delivery: Channel=WEB_SOCKET, Title=Your OTP Code
📤 Created 1 delivery records for notification=notif-123

# User disconnects
📍 Registry: UserID=user-1 unregistered (no more connections)
User user-1 disconnected

# Graceful shutdown
🛑 Shutting down dispatcher workers...
[Worker 0] ✔️ Cleaned up and exited
[Worker 1] ✔️ Cleaned up and exited
[Worker 2] ✔️ Cleaned up and exited
[Worker 3] ✔️ Cleaned up and exited
✅ All workers finished.
```

---

### **Example 8: Error Handling**

```go
// SendNotificationHandle error cases

// Case 1: Invalid JSON
curl -X POST http://localhost:8080/api/v1/send \
  -d 'invalid json'

// Response: 400 Bad Request
// "Invalid request body"

// Case 2: Redis unavailable
// Response: 500 Internal Server Error
// "Failed to publish message"

// Case 3: WebSocket upgrade failure
curl http://localhost:8080/ws?user_id=user-1
// Response: 400 Bad Request
// "User ID is required"

// Case 4: WebSocket write timeout (handled gracefully)
[Worker 0] ⚠️ Send channel timeout for UserID=user-1
// Worker doesn't crash, continues processing next message
```

---

### **Example 9: Registry Data Structure**

```go
// How Registry tracks connections

Registry.clients = {
    "user-1": {
        &Client{UserID: "user-1", SendChan: chan1}: true,
        &Client{UserID: "user-1", SendChan: chan2}: true,  // Tab B
        &Client{UserID: "user-1", SendChan: chan3}: true,  // Mobile
    },
    "user-2": {
        &Client{UserID: "user-2", SendChan: chan4}: true,
    },
}

// When notification arrives for user-1:
// GetClients("user-1") returns [chan1, chan2, chan3]
// Worker sends to all 3 clients simultaneously
// 3x broadcast capability ✅
```

---

### **Example 10: Performance Metrics**

```
Configuration:
- IngestionChan buffer: 1000 messages
- Worker pool: 4 concurrent workers
- WebSocket send timeout: 2 seconds
- Client send channel buffer: 256 messages

Theoretical throughput:
- 1000 messages/batch × 4 workers = 4000 msgs/second
- With Redis bottleneck: ~1000-2000 msgs/second realistic
- Per-user latency: <100ms for online users

Scalability:
- Horizontal: Add more nodes (Redis coordinates)
- Vertical: Increase workers (goroutines)
- Concurrency: Gorilla WebSocket handles 10K+ connections
```

---

## 🔄 Retry Flow (Phase 1 - Future)

```go
// NotificationService.HandleRetry()

// Failed delivery triggers:
// 1. Create retry record with:
//    - Max retries: 3
//    - Initial backoff: 30 seconds
//    - Exponential increase: 30s → 60s → 120s

// Retry scheduler (background job):
// 1. Query: GetPendingRetries(limit=100)
// 2. Check: next_retry_at <= NOW()
// 3. Action: Re-send with UpdateDeliveryStatus()
// 4. Track: IncrementRetryCount()
// 5. Complete: MarkRetryCompleted() if successful

// Status transitions:
// PENDING → SENT (success)
//        → FAILED → PENDING (retry) → SENT
//        → FAILED (max retries exceeded)
```

---

## 📚 Summary

✅ **Real-time**: Online users receive immediately via WebSocket
✅ **Reliable**: Offline messages stored in DB for later retrieval
✅ **Scalable**: Multi-node coordination via Redis
✅ **Concurrent**: Worker pool with goroutines
✅ **Testable**: Clear separation of concerns
✅ **Observable**: Comprehensive logging with indicators

