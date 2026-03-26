package models

import "time"

type Place struct {
	ID          uint64       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       int64        `json:"price"`
	Locality    Locality     `json:"locality"`
	Category    Category     `json:"category"`
	Photos      []PlacePhoto `json:"photos"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type PlacePhoto struct {
	ID       uint64 `json:"id"`
	PlaceID  uint64 `json:"place_id"`
	FilePath string `json:"file_path"`
	IsMain   bool   `json:"is_main"`
}
