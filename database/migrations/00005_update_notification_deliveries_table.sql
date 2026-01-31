-- +goose Up
-- +goose StatementBegin

-- Drop old columns
ALTER TABLE notification_deliveries
DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS content;

-- Add new multi-language columns
ALTER TABLE notification_deliveries
    ADD COLUMN title_vi TEXT NOT NULL DEFAULT '',
    ADD COLUMN title_en TEXT NOT NULL DEFAULT '',
    ADD COLUMN content_vi TEXT NOT NULL DEFAULT '',
    ADD COLUMN content_en TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE notification_deliveries
DROP COLUMN IF EXISTS title_vi,
    DROP COLUMN IF EXISTS title_en,
    DROP COLUMN IF EXISTS content_vi,
    DROP COLUMN IF EXISTS content_en;

ALTER TABLE notification_deliveries
    ADD COLUMN title TEXT,
    ADD COLUMN content TEXT;

-- +goose StatementEnd