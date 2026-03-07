package models

import "time"

type Favorite struct {
	UserID    int64
	PlaceID   int64
	CreatedAt time.Time
}
