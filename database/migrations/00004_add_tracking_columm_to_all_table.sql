-- +goose Up
-- +goose StatementBegin

ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS created_by VARCHAR (255) DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR (255) DEFAULT 'system';

ALTER TABLE notification_deliveries
    ADD COLUMN IF NOT EXISTS created_by VARCHAR (255) DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR (255) DEFAULT 'system';

ALTER TABLE notification_templates
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS created_by VARCHAR (255) DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR (255) DEFAULT 'system';

ALTER TABLE notification_retries
    ADD COLUMN IF NOT EXISTS created_by VARCHAR (255) DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR (255) DEFAULT 'system';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE notifications
DROP
COLUMN IF EXISTS created_by,
    DROP
COLUMN IF EXISTS updated_at,
    DROP
COLUMN IF EXISTS updated_by;

ALTER TABLE notification_deliveries
DROP
COLUMN IF EXISTS created_by,
    DROP
COLUMN IF EXISTS updated_at,
    DROP
COLUMN IF EXISTS updated_by;

ALTER TABLE notification_templates
DROP
COLUMN IF EXISTS created_at,
    DROP
COLUMN IF EXISTS created_by,
    DROP
COLUMN IF EXISTS updated_at,
    DROP
COLUMN IF EXISTS updated_by;

ALTER TABLE notification_retries
DROP
COLUMN IF EXISTS created_by,
    DROP
COLUMN IF EXISTS updated_by;

-- +goose StatementEnd