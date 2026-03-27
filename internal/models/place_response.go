package models

import "time"

type PlaceResponse struct {
	ID          uint64       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       int64        `json:"price"`
	IsLiked     bool         `json:"is_liked"`
	Locality    Locality     `json:"locality"`
	Category    *Category    `json:"category,omitempty"`
	Photos      []PlacePhoto `json:"photos,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}
