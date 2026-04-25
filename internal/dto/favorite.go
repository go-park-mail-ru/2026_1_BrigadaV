package dto

import "time"

type FavoriteResponse struct {
	UserID    uint64    `json:"user_id"`
	PlaceID   uint64    `json:"place_id"`
	CreatedAt time.Time `json:"created_at"`
}
