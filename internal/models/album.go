package models

import "time"

type Album struct {
	ID           uint64
	TripID       uint64
	Name         string
	Description  string
	CoverPhotoID *uint64
	MaxPhotos    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
