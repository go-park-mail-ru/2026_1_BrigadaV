package models

import "time"

type Favorite struct {
	UserID    uint64
	PlaceID   uint64
	CreatedAt time.Time
}
