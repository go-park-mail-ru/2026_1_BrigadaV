-- +goose Up
-- +goose StatementBegin


UPDATE trip_member SET role = 'companion' WHERE role = 'editor';


ALTER TABLE trip_member DROP CONSTRAINT IF EXISTS trip_member_role_check;


ALTER TABLE trip_member ADD CONSTRAINT trip_member_role_check CHECK (role IN ('owner', 'companion', 'viewer'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Возвращаем старый constraint
UPDATE trip_member SET role = 'editor' WHERE role = 'companion';
ALTER TABLE trip_member DROP CONSTRAINT IF EXISTS trip_member_role_check;
ALTER TABLE trip_member ADD CONSTRAINT trip_member_role_check CHECK (role IN ('owner', 'editor', 'viewer'));

-- +goose StatementEnd