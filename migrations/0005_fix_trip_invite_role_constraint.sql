-- +goose Up
-- +goose StatementBegin

-- Переименовать роль 'editor' в 'companion' в существующих записях trip_invite
UPDATE trip_invite SET role = 'companion' WHERE role = 'editor';

-- Пересоздать CHECK constraint с правильными значениями ролей
ALTER TABLE trip_invite DROP CONSTRAINT IF EXISTS trip_invite_role_check;
ALTER TABLE trip_invite ADD CONSTRAINT trip_invite_role_check CHECK (role IN ('companion', 'viewer'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE trip_invite SET role = 'editor' WHERE role = 'companion';
ALTER TABLE trip_invite DROP CONSTRAINT IF EXISTS trip_invite_role_check;
ALTER TABLE trip_invite ADD CONSTRAINT trip_invite_role_check CHECK (role IN ('editor', 'viewer'));

-- +goose StatementEnd
