package models

import "time"

type Favorite struct {
	UserID    uint64    `json:"user_id"`
	PlaceID   uint64    `json:"place_id"`
	CreatedAt time.Time `json:"created_at"`
}
