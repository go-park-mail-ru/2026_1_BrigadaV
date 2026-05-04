package models

import "time"

type Place struct {
	ID          uint64
	Name        string
	Description string
	PhotoURL    string
	Type        string
	LocalityID  *uint64
	CategoryID  *uint64
	Price       int64
	Rating      float64
	ReviewCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Locality Locality
	Category Category
	Photos   []PlacePhoto
}

type PlaceWithRating struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PhotoURL    string   `json:"photo_url"`
	Price       int64    `json:"price"`
	Rating      float64  `json:"rating"`
	ReviewCount int64    `json:"reviewCount"`
	IsLiked     bool     `json:"is_liked"`
	Locality    Locality `json:"locality"`
	Category    *Category `json:"category,omitempty"`
}

type PlaceInTrip struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	PhotoURL    string  `json:"photo_url"`
	Rating      float64 `json:"rating"`
}
