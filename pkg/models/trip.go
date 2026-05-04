package models

import "time"

type Trip struct {
	ID          uint64
	Title       string
	Description string
	Location    *string
	StartDate   *time.Time
	EndDate     *time.Time
	PreviewURL  *string
	CreatedBy   uint64
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TripAttraction struct {
	TripID     uint64
	PlaceID    uint64
	OrderIndex int16
	CreatedAt  time.Time
}
