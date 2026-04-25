package dto

import "time"

type CreateAlbumRequest struct {
	TripID      uint64 `json:"trip_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateAlbumRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type AlbumResponse struct {
	ID           uint64    `json:"id"`
	TripID       uint64    `json:"trip_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	CoverPhotoID *uint64   `json:"cover_photo_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
