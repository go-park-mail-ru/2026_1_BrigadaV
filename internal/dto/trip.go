package dto

import (
	"guidely-app/internal/models"
	"time"
)

type TripResponse struct {
	ID        uint64     `json:"id"`
	Title     string     `json:"title"`
	Location  *string    `json:"location,omitempty"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
	Preview   *string    `json:"preview,omitempty"`
}

type CreateTripRequest struct {
	Title     string  `json:"title"`
	Location  *string `json:"location,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
	Preview   *string `json:"preview,omitempty"`
	IsPublic  bool    `json:"is_public"`
}

type CreateTripResponse struct {
	ID      uint64  `json:"id"`
	Preview *string `json:"preview,omitempty"`
}

type UpdateTripRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Location    *string `json:"location,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	Preview     *string `json:"preview,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}

type TripDetailsResponse struct {
	ID          uint64               `json:"id"`
	Title       string               `json:"title"`
	Location    *string              `json:"location,omitempty"`
	StartDate   *time.Time           `json:"startDate,omitempty"`
	EndDate     *time.Time           `json:"endDate,omitempty"`
	Preview     *string              `json:"preview,omitempty"`
	Attractions []models.PlaceInTrip `json:"attractions"`
}
