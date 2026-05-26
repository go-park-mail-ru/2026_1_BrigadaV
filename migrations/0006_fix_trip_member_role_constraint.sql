-- +goose Up
-- +goose StatementBegin

-- Обновить CHECK constraint: добавить 'companion' (данные были переименованы в 0004)
ALTER TABLE trip_member DROP CONSTRAINT IF EXISTS trip_member_role_check;
ALTER TABLE trip_member ADD CONSTRAINT trip_member_role_check
    CHECK (role IN ('owner', 'editor', 'companion', 'viewer'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE trip_member DROP CONSTRAINT IF EXISTS trip_member_role_check;
ALTER TABLE trip_member ADD CONSTRAINT trip_member_role_check
    CHECK (role IN ('owner', 'editor', 'viewer'));

-- +goose StatementEnd
