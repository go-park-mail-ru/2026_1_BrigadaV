package dto

import "time"

type CreateTripRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	IsPublic    bool    `json:"is_public"`
}

type UpdateTripRequest struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}

type TripResponse struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedBy   uint64     `json:"created_by"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
