package models

import "time"

type PlacePhoto struct {
	ID        int64     `json:"id"`
	PlaceID   int64     `json:"place_id"`
	FilePath  string    `json:"file_path"`
	IsMain    bool      `json:"is_main"`
	CreatedAt time.Time `json:"created_at"`
}
