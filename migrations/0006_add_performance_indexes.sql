-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Индексы для place
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_place_rating_reviews ON place (rating DESC, review_count DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_place_name_trgm ON place USING gin (name gin_trgm_ops);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_place_description_trgm ON place USING gin (description gin_trgm_ops);

-- Индекс для review
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_review_place_created ON review (place_id, created_at DESC);

-- Индексы для trip_member и trip_attractions
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trip_member_user ON trip_member (user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trip_attractions_trip ON trip_attractions (trip_id);

-- Индекс для album_photo
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_album_photo_album ON album_photo (album_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_place_rating_reviews;
DROP INDEX IF EXISTS idx_place_name_trgm;
DROP INDEX IF EXISTS idx_place_description_trgm;
DROP INDEX IF EXISTS idx_review_place_created;
DROP INDEX IF EXISTS idx_trip_member_user;
DROP INDEX IF EXISTS idx_trip_attractions_trip;
DROP INDEX IF EXISTS idx_album_photo_album;
-- +goose StatementEnd