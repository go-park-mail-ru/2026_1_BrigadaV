-- +goose Up
-- +goose StatementBegin
UPDATE trip_member SET role = 'companion' WHERE role = 'editor';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE trip_member SET role = 'editor' WHERE role = 'companion';
-- +goose StatementEnd