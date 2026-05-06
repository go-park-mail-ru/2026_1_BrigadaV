package dto

import "time"

type PlaceResponse struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	PhotoURL    string          `json:"photo_url"`
	Price       int64           `json:"price"`
	IsLiked     bool            `json:"is_liked"`
	Latitude    *float64        `json:"latitude,omitempty"`
	Longitude   *float64        `json:"longitude,omitempty"`
	Locality    LocalityDTO     `json:"locality"`
	Category    *CategoryDTO    `json:"category,omitempty"`
	Photos      []PlacePhotoDTO `json:"photos,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type LocalityDTO struct {
	ID        uint64   `json:"id"`
	Name      string   `json:"name"`
	Country   string   `json:"country"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type CategoryDTO struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PlacePhotoDTO struct {
	ID       uint64 `json:"id"`
	PlaceID  uint64 `json:"place_id"`
	FilePath string `json:"file_path"`
	IsMain   bool   `json:"is_main"`
}
