-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications
    ADD COLUMN channels TEXT[] DEFAULT ARRAY['WEB_SOCKET'];

CREATE INDEX idx_noti_channels ON notifications USING GIN (channels);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_noti_channels;
ALTER TABLE notifications
DROP
COLUMN channels;
-- +goose StatementEnd
