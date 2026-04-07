CREATE TABLE IF NOT EXISTS album_photo (
    album_id BIGINT NOT NULL REFERENCES album(id) ON DELETE CASCADE,
    photo_id BIGINT NOT NULL REFERENCES photo(id) ON DELETE CASCADE,
    order_index SMALLINT NOT NULL DEFAULT 0 CHECK (order_index >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (album_id, photo_id)
);