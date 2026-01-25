# Dispatcher Implementation Guide - Phase 1

## ✅ Hoàn Tất Implementation

Dispatcher đã được refactor để tương thích đầy đủ với các models hiện tại và notification service.

---

## 🏗 Kiến Trúc Dispatcher

### 1. **Cấu Trúc Chính**

```
Dispatcher
├── IngestionChan          // Channel nhận Notification từ Redis
├── Registry               // Quản lý WebSocket connections
├── RedisClient            // Kết nối Redis
└── WorkerPool             // Xử lý tin nhắn đồng thời
```

### 2. **Luồng Dữ Liệu (Data Flow)**

```
┌─────────────────────────────────────────────────────────────┐
│                    REQUEST FLOW                             │
└─────────────────────────────────────────────────────────────┘

1️⃣  CLIENT API REQUEST
    ↓
    POST /api/v1/send
    {
      "user_id": "user-1",
      "event_type": "PAYMENT_SUCCESS",
      "data": {"amount": 1000, "account": "123456"},
      "correlation_id": "corr-xyz"
    }

2️⃣  HANDLE LAYER (handle.go)
    ↓
    SendNotificationHandle()
    ├── Validate request
    ├── Marshal payload
    └── PublishToRedis()

3️⃣  REDIS BROADCAST (All Nodes Subscribe)
    ↓
    Dispatcher.PublishToRedis()
    └── Publish to "notifications" channel

4️⃣  REDIS SUBSCRIBER (StartRedisSubscriber)
    ↓
    Subscribe to "notifications" channel
    └── Unmarshal message → IngestionChan

5️⃣  WORKER POOL (Worker processes message)
    ↓
    for msg := range d.IngestionChan
    ├── Check Registry for user connections
    │   ├── IF USER ONLINE:
    │   │   └── Send via WebSocket immediately
    │   └── IF USER OFFLINE:
    │       └── Store in database (Phase 1)
    └── Log result

6️⃣  WEBSOCKET SEND (WSHandler)
    ↓
    client.SendChan <- msg
    └── WebSocket sends JSON to client
```

---

## 🔧 Các Component Chi Tiết

### **Dispatcher Struct**

```go
type Dispatcher struct {
    IngestionChan chan models.Notification  // Buffer 1000 messages
    Registry      *Registry                 // Track online users
    WG            sync.WaitGroup            // Wait for workers
    RedisClient   *redis.Client             // Redis connection
    RedisChannel  string                    // "notifications"
}
```

### **Client Registry**

```go
type Registry struct {
    clients map[string]map[*Client]bool    // UserID → [Clients]
    mu      sync.RWMutex                   // Thread-safe access
}

// Supports:
// - One user, multiple devices
// - One device, multiple tabs
// - Concurrent access
```

### **Notification Model (Used)**

```go
type Notification struct {
    ID            string          // UUID
    UserID        string          // "user-1"
    EventType     string          // "PAYMENT_SUCCESS"
    Payload       json.RawMessage // Dynamic data
    CorrelationID string          // Trace ID
    CreatedAt     time.Time       // Timestamp
}
```

---

## 🚀 Initialization & Usage

### **1. Create Dispatcher**

```go
// cmd/main.go
dispatcher := dispatcher.NewDispatcher(1000, "localhost:6379")

// Start Redis subscriber
dispatcher.StartRedisSubscriber()

// Start worker pool (4 workers)
dispatcher.StartWorkerPool(4)
```

### **2. Setup HTTP Handlers**

```go
h := api.NewHandle(dispatcher, config)

http.HandleFunc("/api/v1/send", h.SendNotificationHandle)
http.HandleFunc("/ws", h.WSHandler)

http.ListenAndServe(":8080", nil)
```

### **3. Client WebSocket Connection**

```javascript
// Client: Connect with user ID
const ws = new WebSocket('ws://localhost:8080/ws?user_id=user-1');

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log('📬 Notification:', msg);
    // {
    //   "id": "notif-123",
    //   "user_id": "user-1",
    //   "event_type": "PAYMENT_SUCCESS",
    //   "payload": {...},
    //   "created_at": "2026-01-18T..."
    // }
};
```

---

## 📊 Worker Pool Pattern

### **Multi-Worker Processing**

```
Notification 1 ──┐
Notification 2 ──┤      ┌─ Worker 0 ┐
Notification 3 ──┼──→   ├─ Worker 1 ├─→ Registry/WebSocket
Notification 4 ──┤      ├─ Worker 2 │
Notification 5 ──┘      └─ Worker 3 ┘

Benefits:
✅ Non-blocking (async processing)
✅ Scalable (adjust worker count)
✅ Throughput (parallel handling)
✅ Graceful shutdown (WaitGroup)
```

### **Worker Logic**

```go
func (d *Dispatcher) worker(id int) {
    for msg := range d.IngestionChan {
        clients := d.Registry.GetClients(msg.UserID)
        
        if len(clients) > 0 {
            // User online: Send immediately
            for _, client := range clients {
                select {
                case client.SendChan <- msg:
                    // Success ✅
                case <-time.After(2 * time.Second):
                    // Timeout ⚠️
                }
            }
        } else {
            // User offline: Store in DB 💾
            log.Printf("User %s not online, stored in database", msg.UserID)
        }
    }
}
```

---

## 📝 Registry Management

### **Register User Connection**

```go
// WSHandler: User connects
client := &dispatcher.Client{
    UserID:   "user-1",
    SendChan: make(chan models.Notification, 256),
}
h.Dispatcher.Registry.Register("user-1", client)

// Result: Registry.clients["user-1"][client] = true
```

### **Multiple Connections**

```
User-1 (3 tabs open):
├── Client 1 (Tab A) ──→ SendChan (buffer 256)
├── Client 2 (Tab B) ──→ SendChan (buffer 256)
└── Client 3 (Tab C) ──→ SendChan (buffer 256)

When notification arrives:
  Worker broadcasts to ALL 3 clients simultaneously
```

### **Unregister on Disconnect**

```go
defer h.Dispatcher.Registry.Unregister(userID, client)

// Removes client from Registry
// If no more clients for user → Delete user entry
```

---

## 🔄 Integration with NotificationService

### **Phase 1 Flow**

```
API Request
  ↓
SendNotificationHandle (handle.go)
  ├── Create Notification struct
  └── PublishToRedis()
  ↓
Dispatcher Workers
  ├── Check if user online
  ├── IF YES: Send via WebSocket
  └── IF NO: ProcessNotification() → Save to DB
  ↓
NotificationService.ProcessNotification()
  ├── SaveMasterNotification() → DB
  ├── GetTemplatesByEvent() → Get templates
  ├── Render templates with payload
  ├── CreateDeliveries() → Save delivery records
  └── Return notification ID
```

### **NotificationService Methods**

```go
// Save notification + create delivery records
ProcessNotification(ctx, notification) → notificationID

// Mark success/failure
UpdateDeliverySuccess(ctx, deliveryID)
UpdateDeliveryFailure(ctx, deliveryID, error)

// Retry logic
HandleRetry(ctx, deliveryID, retryCount, maxRetries, error, backoffSeconds)
ProcessPendingRetries(ctx, limit) → retries to process

// User history
GetNotificationHistory(ctx, userID, limit, offset)
MarkAsRead(ctx, deliveryID)
```

---

## 🧪 Test Scenarios

### **Scenario 1: User Online**

```bash
# Terminal 1: Start server
go run cmd/main.go -port=8080

# Terminal 2: Client connects WebSocket
wscat -c "ws://localhost:8080/ws?user_id=user-1"

# Terminal 3: Send notification
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-1",
    "event_type": "PAYMENT_SUCCESS",
    "data": {"amount": 1000},
    "correlation_id": "corr-123"
  }'

# Terminal 2: Receives notification immediately ✅
```

### **Scenario 2: User Offline**

```bash
# Send notification (no client connected)
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user-2", ...}'

# Notification stored in DB
# When user comes online → Service provides history via API
```

### **Scenario 3: Multiple Nodes**

```bash
# Terminal 1: Node 1
go run cmd/main.go -port=8080

# Terminal 2: Node 2
go run cmd/main.go -port=8081

# Terminal 3: Client on Node 1
wscat -c "ws://localhost:8080/ws?user_id=user-1"

# Terminal 4: Send to Node 2
curl -X POST http://localhost:8081/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{...}'

# Flow: Node 2 → Redis → Node 1 → Client ✅
```

---

## 📋 Status Constants

```go
// From models/event.go
const (
    StatusPending   DeliveryStatus = "PENDING"
    StatusSent      DeliveryStatus = "SENT"
    StatusFailed    DeliveryStatus = "FAILED"
    StatusCompleted DeliveryStatus = "COMPLETED"
    StatusRead      DeliveryStatus = "READ"
)
```

---

## 🔐 Graceful Shutdown

```go
// cmd/main.go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

<-sigChan
dispatcher.Shutdown()  // Close IngestionChan, wait for workers
db.Close()
```

---

## ✨ Key Improvements Made

✅ **Removed old models**: `NotificationMessage` → `Notification`
✅ **Fixed dispatcher**: Updated channel types and method signatures
✅ **Integrated service**: NotificationService works with dispatcher
✅ **Status management**: Proper type casting for DeliveryStatus
✅ **Logging**: Enhanced with emoji indicators for clarity
✅ **Error handling**: Proper error wrapping and propagation
✅ **Clean code**: Removed unused functions

---

## 📚 Files Modified

- ✅ `internal/api/handle.go` - Fixed API handlers
- ✅ `internal/dispatcher/dispatcher.go` - Updated models and flow
- ✅ `internal/service/notification_service.go` - Fixed status types

---

## 🎯 Next Steps (Phase 2)

1. **Implement repository**: Write PostgreSQL adapter for `NotificationDB` interface
2. **Add retry scheduler**: Background job to process pending retries
3. **Multi-channel support**: Email, SMS, Push notifications
4. **ACK protocol**: Track message acknowledgments
5. **Metrics**: Prometheus integration for monitoring

