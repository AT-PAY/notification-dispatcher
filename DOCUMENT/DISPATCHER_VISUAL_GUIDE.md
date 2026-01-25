# 🎯 Dispatcher System - Visual Flow Guide

## 📡 Message Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1️⃣  REST API Client              2️⃣  WebSocket Client              │
│  ┌──────────────────┐             ┌──────────────────┐              │
│  │ POST /api/v1/send│             │ ws://...?user=U1 │              │
│  │                  │             │                  │              │
│  │ {                │             │ Persistent conn  │              │
│  │  user_id: "U1"   │             │                  │              │
│  │  event_type: "X" │             │ Receives         │              │
│  │  data: {...}     │             │ notifications    │              │
│  │ }                │             │ in real-time     │              │
│  └────────┬─────────┘             └────────┬─────────┘              │
│           │                                │                        │
└───────────┼────────────────────────────────┼────────────────────────┘
            │                                │
            ▼                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      API HANDLER LAYER                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  api.Handle.SendNotificationHandle()                                │
│  ├── Decode JSON                                                    │
│  ├── Validate                                                       │
│  ├── Create Notification{                                           │
│  │     UserID: "U1",                                                │
│  │     EventType: "PAYMENT_SUCCESS",                                │
│  │     Payload: json.RawMessage,                                    │
│  │     CreatedAt: now()                                             │
│  │   }                                                              │
│  └── Dispatcher.PublishToRedis()   ──────────┐                      │
│                                              │                      │
│  api.Handle.WSHandler()                      │                      │
│  ├── Upgrade HTTP to WebSocket               │                      │
│  ├── Create Client{UserID, SendChan}         │                      │
│  └── Registry.Register(userID, client)   ◄───┼─────────────┐        │
│                                              │             │        │
└──────────────────────────────────────────────┼─────────────┼────────┘
                                               │             │
                    ┌──────────────────────────┘             │
                    │                                        │
                    ▼                                        │
    ┌──────────────────────────────┐         Registry        │
    │   REDIS PUB/SUB              │         ─────────       │
    │                              │         {               │
    │  Channel: "notifications"    │         "U1": [         │
    │                              │           client1,      │
    │  Payload: Notification{}     │           client2       │
    │                              │         ]               │
    │  All nodes subscribed ◄──────┼─────────}               │
    │                              │                         │
    └──────────────────────────────┘                         │
                    │                                        │
                    │ Broadcast to all nodes                 │
                    │ (multi-node coordination)              │
                    │                                        │
        ┌───────────┼───────────┐                            │
        │           │           │                            │
        ▼           ▼           ▼                            │
    ┌────────┐ ┌────────┐ ┌────────┐                         │
    │ Node 1 │ │ Node 2 │ │ Node 3 │                         │
    └───┬────┘ └───┬────┘ └───┬────┘                         │
        │          │          │                              │
        ▼          ▼          ▼                              │
    ┌──────────────────────────────────────────┐             │
    │  StartRedisSubscriber() on each node     │             │
    │  ├── Receive from Redis                  │             │
    │  ├── Unmarshal JSON → Notification       │             │
    │  └── Put into d.IngestionChan ──────┐    │             │
    └──────────────────────────────────────────┘             │
                                              │              │
                                              ▼              │
                                    ┌────────────────────┐   │
                                    │  IngestionChan     │   │
                                    │  (buffered queue)  │   │
                                    │  capacity: 1000    │   │
                                    └───────┬────────────┘   │
                                            │                │
                    ┌───────────────────────┼────────────────┼────────────┐
                    │                       │                │            │
                    ▼                       ▼                ▼            ▼
            ┌───────────────┐       ┌───────────────┐      ┌──────────────────┐
            │   Worker 0    │       │   Worker 1    │      │   Worker 2/3     │
            │               │       │               │      │                  │
            │ consume msg   │       │ consume msg   │      │ (similar)        │
            │ from chan     │       │ from chan     │      │                  │
            └────┬──────────┘       └────┬──────────┘      └────┬─────────────┘
                 │                       │                      │
                 ├───────┬───────────────┴──────────────────────┤
                 │       │                                      │
                 ▼       ▼                                      ▼
            ┌─────────────────────────────────────────────────────────┐
            │  Check: Is user online on this node?                    │
            ├─────────────────────────────────────────────────────────┤
            │  GetClients(userID) → lookup in Registry                │
            └──────────┬───────────────────────────────┬──────────────┘
                       │                               │
         ┌─────────────┘                               └──────────────┐
         │                                                            │
         ▼ YES (User Online)                          ▼ NO (Offline)  │
    ┌──────────────────────┐                    ┌─────────────────────┐
    │ For each client:     │                    │ Store in Database   │
    │ ├── client.SendChan  │                    │ via                 │
    │ ├── <- msg           │                    │ NotificationService │
    │ ├── Timeout: 2sec    │                    │                     │
    │ └── Log: ✅ Sent     │                    │ ProcessNotification │
    └────────┬─────────────┘                    │ ├── Save master     │
             │                                  │ ├── Get templates   │
             │                                  │ ├── Render content  │
             │                                  │ ├── Create delivery │
             │                                  │ └── Save to DB      │
             │                                  └─────────┬───────────┘
             │                                            │
             └────────────────┬───────────────────────────┘
                              │
                              ▼
                    ┌─────────────────────────┐
                    │   WSHandler             │
                    │  ├── Receive from       │
                    │  │   client.SendChan    │
                    │  ├── Marshal JSON       │
                    │  └── WebSocket.Write()  │
                    │      ✅ Send to client  │
                    └─────────────────────────┘
                              │
                              ▼
                    ┌─────────────────────────┐
                    │  CLIENT (Browser/App)   │
                    │  ├── onmessage()        │
                    │  ├── Parse JSON         │
                    │  └── Display            │
                    │      notification 📬    │
                    └─────────────────────────┘
```

---

## 🔀 Multi-Node Coordination Example

```
┌──────────────────────────────────────────────────────────────────────┐
│                        CLUSTER SETUP                                 │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Node 1 (Port 8080)           Node 2 (Port 8081)                     │
│  ┌──────────────────────┐     ┌──────────────────────┐               │
│  │ Dispatcher Instance  │     │ Dispatcher Instance  │               │
│  ├──────────────────────┤     ├──────────────────────┤               │
│  │ Registry:            │     │ Registry:            │               │
│  │ - user-1 (1 conn)    │     │ - user-2 (1 conn)    │               │
│  │                      │     │                      │               │
│  │ Workers: 4           │     │ Workers: 4           │               │
│  └──────────┬───────────┘     └──────────┬───────────┘               │
│             │                            │                           │
│             │ Subscribe                  │ Subscribe                 │
│             └────────────┬───────────────┘                           │
│                          │                                           │
│                          ▼                                           │
│                  ┌──────────────────┐                                │
│                  │  REDIS           │                                │
│                  │                  │                                │
│                  │ Channel:         │                                │
│                  │ notifications    │                                │
│                  │                  │                                │
│                  │ Both nodes       │                                │
│                  │ connected        │                                │
│                  └──────────────────┘                                │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘

Scenario: User connected to Node 1, notification sent via Node 2

Step 1: Node 2 receives HTTP request
        POST /api/v1/send (node2:8081)
        User-1 notification

Step 2: Node 2 publishes to Redis
        Dispatcher.PublishToRedis() 
        → "notifications" channel

Step 3: Both nodes receive from Redis
        Node 1: ✅ I have user-1 connected
        Node 2: ❌ I don't have user-1

Step 4: Node 1 worker processes
        Registry.GetClients("user-1") → [client]
        Send to WebSocket

Step 5: Client receives notification
        From Node 1 ✅ 
        (even though request was sent to Node 2)

Result: Seamless multi-node delivery!
```

---

## 🔄 Retry Mechanism (Phase 1 Future)

```
┌────────────────────────────────────────────────────────────┐
│              DELIVERY FAILURE → RETRY FLOW                 │
└────────────────────────────────────────────────────────────┘

Initial Delivery Attempt:
Notification → Worker → User Offline
                       ↓
                    Database:
                    - notification_deliveries
                      status: PENDING
                    - notification_retries
                      retry_count: 0
                      max_retries: 3
                      last_error: "User offline"
                      next_retry_at: NOW() + 30s

                       ↓
                   [WAIT 30 SECONDS]
                       ↓

Retry Attempt 1 (30s later):
GetPendingRetries()
├── Query: next_retry_at <= NOW()
├── Result: Find retry record
└── Action: IncrementRetryCount()
           next_retry_at: NOW() + 60s
           
If user still offline:
├── Update error
└── Schedule next retry (exponential backoff)

Exponential Backoff:
Attempt 1: 30 seconds
Attempt 2: 60 seconds  
Attempt 3: 120 seconds

After 3 failed attempts:
├── status: FAILED
├── Mark delivery as expired
└── User gets history on login
```

---

## 📊 Status Transitions

```
NOTIFICATION LIFECYCLE:

┌─────────────────────────────────────────────────────┐
│  NotificationDelivery Status Flow                   │
├─────────────────────────────────────────────────────┤

Creation:
  NEW → PENDING
        ↓
User Online:
  PENDING → SENT (WebSocket delivery success)
        ↓
User reads:
  SENT → READ
  
User Offline:
  PENDING → (Wait for retry)
        ↓
After Retries:
  PENDING → SENT (Success) or FAILED (Max retries)
        ↓
Final State:
  READ or FAILED

Status Diagram:

           ┌─────────────────────┐
           │    NEW/PENDING      │
           └──────────┬──────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
    [Online]      [Offline]    [Error]
        │             │             │
        ▼             ▼             ▼
    ┌────────┐  ┌────────────┐ ┌────────┐
    │ SENT   │  │Retry Queue │ │FAILED  │
    └───┬────┘  └─────┬──────┘ └────────┘
        │              │
        ▼              ▼
    ┌────────┐    ┌────────┐
    │ READ   │    │ SENT   │
    └────────┘    └───┬────┘
                      │
                      ▼
                   ┌────────┐
                   │ READ   │
                   └────────┘
```

---

## 💼 Production Readiness Checklist

```
✅ IMPLEMENTED (Phase 1)
├── Real-time WebSocket delivery
├── Multi-node coordination via Redis
├── Worker pool concurrency (4 workers)
├── Registry for connection tracking
├── Offline database storage structure
├── Graceful shutdown handling
├── Error handling and timeouts
└── Comprehensive logging

🟡 TODO (Phase 2)
├── Retry scheduler (background job)
├── Email/SMS multi-channel support
├── Push notification integration
├── Metrics & monitoring
├── Rate limiting
├── Authentication & security
└── Performance optimization

🔴 NOT IN SCOPE
├── Distributed transactions (Redis handles state)
├── Message ordering guarantees
├── Encryption (implement in API Gateway)
└── Custom message routing rules
```

---

## 🎯 Performance Targets

```
Configuration:
├── IngestionChan buffer: 1000 messages
├── Worker pool: 4 concurrent goroutines
├── WebSocket send timeout: 2 seconds
├── Client send buffer: 256 messages per connection
└── Redis: Standard configuration

Expected Performance:
├── Throughput: 1000-2000 messages/second
├── Latency (online): <100ms
├── Concurrent connections: 10,000+ per node
├── Memory per connection: ~64KB
└── Scalability: Linear with worker count

Bottlenecks:
├── Redis throughput (pub/sub)
├── Database persistence (if offline)
├── Network I/O (WebSocket)
└── CPU (JSON marshaling)
```

---

## 🛠 Troubleshooting Guide

```
Issue: "User not online, message stored in database"
└── Expected behavior when user disconnected
└── Solution: Check offline history endpoint

Issue: "Send channel timeout"
└── WebSocket send blocked for 2+ seconds
└── Likely: Network issue or overloaded client
└── Solution: Check network, increase buffer size

Issue: "Error unmarshaling Redis message"
└── Redis payload corrupted or wrong format
└── Solution: Verify Redis connection, check Notification struct

Issue: "User ID is required"
└── WebSocket connection missing ?user_id parameter
└── Solution: Use ws://...?user_id=YOUR_USER_ID

Issue: "Method not allowed"
└── Using wrong HTTP method on /api/v1/send
└── Solution: Use POST, not GET

Issue: "Failed to upgrade to websocket"
└── HTTP connection can't upgrade to WebSocket
└── Solution: Ensure proper headers in client request
```

---

## 📚 Component Summary Table

| Component | Purpose | Status |
|-----------|---------|--------|
| **Handle** | HTTP/WebSocket routing | ✅ Ready |
| **Dispatcher** | Message orchestration | ✅ Ready |
| **Registry** | Connection tracking | ✅ Ready |
| **Worker Pool** | Concurrent processing | ✅ Ready |
| **NotificationService** | Business logic | ✅ Ready |
| **Redis** | Pub/Sub coordination | ✅ Ready |
| **Database** | Persistence | 🟡 Interface ready |
| **Retry Scheduler** | Auto-retry | 🟡 Next phase |
| **Metrics** | Monitoring | 🟡 Next phase |

---

Generated: 2026-01-18 | Version: 1.0 | Phase: 1 (Core)
