package models

import "time"

type PlacePhoto struct {
	ID        uint64
	PlaceID   uint64
	PhotoID   uint64
	IsMain    bool
	CreatedAt time.Time
	Photo     Photo
}
