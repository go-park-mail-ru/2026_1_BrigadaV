package models

import "time"

type Album struct {
	ID           uint64    `json:"id"`
	TripID       uint64    `json:"trip_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	CoverPhotoID *uint64   `json:"cover_photo_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
