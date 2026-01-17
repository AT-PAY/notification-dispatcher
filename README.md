# 🚀 Notification Dispatcher System

Hệ thống điều phối thông báo thời gian thực (Real-time Notification) hiệu năng cao, hỗ trợ đa thiết bị, đa node và kiến trúc hướng sự kiện (Event-driven).

---

## 📌 Mục đích dự án
Xây dựng một lớp trung gian (Middleware) tin cậy để đẩy thông báo từ các service nghiệp vụ (Banking, E-commerce,...) đến người dùng cuối thông qua nhiều kênh, đảm bảo tính sẵn sàng cao và khả năng mở rộng ngang (Horizontal Scaling).

---

## 🏗 Kiến trúc hệ thống
Hệ thống được thiết kế theo mô hình Worker Pool kết hợp với Message Bus để đồng bộ hóa trạng thái giữa các node.

### Các thành phần chính:
* **Inbound Handlers:** Hỗ trợ nhận tin nhắn qua cả REST API (HTTP) và Kafka Consumer.
* **Redis Pub/Sub:** Đóng vai trò "Cầu nối" (Orchestrator) giúp các node Go chia sẻ tin nhắn với nhau.
* **Worker Pool:** Xử lý tin nhắn đồng thời (Concurrency) để không gây nghẽn hệ thống.
* **WebSocket Registry:** Quản lý danh sách kết nối sống của người dùng (hỗ trợ một người dùng mở nhiều tab/thiết bị cùng lúc).

### Luồng đi của tin nhắn (Data Flow):
1. **Tiếp nhận:** Tin nhắn đến từ API hoặc Kafka.
2. **Phát tán (Broadcast):** Node tiếp nhận sẽ đẩy tin nhắn lên Redis Channel.
3. **Thu thập:** Tất cả các Node trong cụm nhận tin từ Redis và đẩy vào IngestionChan nội bộ.
4. **Điều phối:** Worker nhặt tin, kiểm tra trong Registry xem User có đang kết nối ở Node này không.
5. **Gửi:** Nếu có, đẩy tin xuống qua kết nối WebSocket tương ứng.

---

## 🛠 Công nghệ sử dụng
* **Language:** Go (Golang)
* **Message Broker:** Apache Kafka (KRaft mode)
* **In-memory DB/PubSub:** Redis
* **Communication:** WebSocket (Gorilla WebSocket)
* **Deployment:** Docker & Docker Compose

---

## 🚀 Cách khởi chạy

### 1. Khởi động hạ tầng (Kafka & Redis)
```bash
docker-compose up -d
```

### 2. Chạy các Node ứng dụng
Mở các terminal khác nhau để giả lập nhiều node:

```bash
# Chạy Node 1 tại port 8080
go run ./cmd/main.go -port=8080

# Chạy Node 2 tại port 8081
go run ./cmd/main.go -port=8081
```

### 3. Test gửi thông báo

**Qua API:**
```bash
curl -X POST http://localhost:8080/api/v1/send -d '{"user_id":"user-1","message":"Hello"}'
```

**Qua Kafka:** Push JSON vào topic `notification_topic`.

---

## 📅 Lộ trình phát triển (TODO)

### 🔴 Phase 1: Persistence & Reliability (Ưu tiên cao)
- [ ] Offline Storage: Tích hợp PostgreSQL/MongoDB để lưu tin nhắn khi User không online.
- [ ] Message History: API lấy lại lịch sử thông báo.
- [ ] Retry Mechanism: Tự động gửi lại tin nhắn nếu gặp lỗi mạng.

### 🟡 Phase 2: Multi-channel & Delivery
- [ ] Email/SMS Provider: Tích hợp SendGrid (Email) và Twilio (SMS).
- [ ] Push Notification: Hỗ trợ Firebase Cloud Messaging (FCM).
- [ ] Acknowledgment (Ack): Cơ chế xác nhận từ phía Client để cập nhật trạng thái "Đã nhận/Đã đọc".

### 🟢 Phase 3: Monitoring & Security
- [ ] Metrics: Tích hợp Prometheus và Grafana để theo dõi số lượng kết nối và latency.
- [ ] Authentication: Bảo mật kết nối WebSocket bằng JWT thông qua API Gateway.
- [ ] Rate Limiting: Chống spam thông báo cho từng User.

---

## 👥 Đóng góp
Dự án đang trong quá trình phát triển để đạt chuẩn Banking-grade. Mọi ý kiến đóng góp vui lòng tạo Issue hoặc Pull Request.
