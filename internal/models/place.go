package models

import "time"

type Place struct {
	ID          uint64
	Name        string
	Description string
	LocalityID  *uint64
	CategoryID  *uint64
	Price       int64
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Locality   Locality   `json:"locality,omitempty"`
  Category   Category   `json:"category,omitempty"`
  Photos     []PlacePhoto `json:"photos,omitempty"`
}

type PlaceWithRating struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       int64   `json:"price"`
	Rating      float64 `json:"rating"`
	ReviewCount int64   `json:"reviewCount"`
	IsLiked     bool    `json:"is_liked"`
}

type PlaceInTrip struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Rating      float64 `json:"rating"`
	Image       *string `json:"image,omitempty"`
}
