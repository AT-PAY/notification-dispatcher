-- +goose Up
-- +goose StatementBegin
CREATE TABLE notification_templates
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type       VARCHAR(50) NOT NULL,
    channel          VARCHAR(20) NOT NULL,
    language         VARCHAR(10)      DEFAULT 'vi',
    title_template   TEXT        NOT NULL,
    content_template TEXT        NOT NULL,
    UNIQUE (event_type, channel, language)
);

CREATE TABLE notifications
(
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        VARCHAR(50) NOT NULL,
    event_type     VARCHAR(50) NOT NULL,
    payload        JSONB       NOT NULL,
    correlation_id VARCHAR(100),
    created_at     TIMESTAMP        DEFAULT NOW()
);

CREATE TABLE notification_deliveries
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID REFERENCES notifications (id),
    channel         VARCHAR(20) NOT NULL,
    title           TEXT,
    content         TEXT,
    status          VARCHAR(20)      DEFAULT 'PENDING',
    error_detail    TEXT,
    sent_at         TIMESTAMP,
    read_at         TIMESTAMP,
    created_at      TIMESTAMP        DEFAULT NOW()
);

CREATE TABLE notification_retries
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_id   UUID REFERENCES notification_deliveries (id),
    retry_count   INT              DEFAULT 0,
    max_retries   INT NOT NULL,
    last_error    TEXT,
    next_retry_at TIMESTAMP,
    status        VARCHAR(20)      DEFAULT 'PENDING',
    created_at    TIMESTAMP        DEFAULT NOW(),
    updated_at    TIMESTAMP        DEFAULT NOW()
);

CREATE INDEX idx_delivery_noti_id ON notification_deliveries (notification_id);
CREATE INDEX idx_noti_user_id ON notifications (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notification_retries;
DROP TABLE IF EXISTS notification_deliveries;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_templates;
-- +goose StatementEnd