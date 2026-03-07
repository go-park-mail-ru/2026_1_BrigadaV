package models

import "time"

type Place struct {
	ID          int64
	Name        string
	Description string
	CityID      int64
	CategoryID  *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PlaceWithDetails struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	City        *City        `json:"city,omitempty"`
	Category    *Category    `json:"category,omitempty"`
	Photos      []PlacePhoto `json:"photos,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}
