package models

import "time"

type Review struct {
	ID        uint64
	UserID    uint64
	PlaceID   uint64
	Rating    int16
	Comment   string
	VisitDate *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
