//go:generate easyjson -all place.go

package dto

import "time"

//easyjson:json
type PlaceResponse struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	PhotoURL    string          `json:"photo_url"`
	Price       int64           `json:"price"`
	IsLiked     bool            `json:"is_liked"`
	Rating      float64         `json:"rating"`
	ReviewCount int             `json:"reviewCount"`
	Latitude    *float64        `json:"latitude,omitempty"`
	Longitude   *float64        `json:"longitude,omitempty"`
	Locality    LocalityDTO     `json:"locality"`
	Category    *CategoryDTO    `json:"category,omitempty"`
	Photos      []PlacePhotoDTO `json:"photos,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

//easyjson:json
type LocalityDTO struct {
	ID        uint64   `json:"id"`
	Name      string   `json:"name"`
	Country   string   `json:"country"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

//easyjson:json
type CategoryDTO struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

//easyjson:json
type PlacePhotoDTO struct {
	ID       uint64 `json:"id"`
	PlaceID  uint64 `json:"place_id"`
	FilePath string `json:"file_path"`
	IsMain   bool   `json:"is_main"`
}
