-- +goose Up
-- +goose StatementBegin

ALTER TABLE "user"
    ADD COLUMN IF NOT EXISTS yandex_id TEXT UNIQUE;

ALTER TABLE "user"
    ALTER COLUMN password_hash SET DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE "user"
    DROP COLUMN IF EXISTS yandex_id;

ALTER TABLE "user"
    ALTER COLUMN password_hash DROP DEFAULT;

-- +goose StatementEnd
