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

//easyjson:json
type PlaceResponseList []PlaceResponse

// PlaceWithRatingResponse используется в GetDetails — совпадает с models.PlaceWithRating
// но сериализуется через easyjson.
//easyjson:json
type PlaceWithRatingResponse struct {
	ID          uint64   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PhotoURL    string   `json:"photo_url"`
	Price       int64    `json:"price"`
	Rating      float64  `json:"rating"`
	ReviewCount int64    `json:"reviewCount"`
	IsLiked     bool     `json:"is_liked"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

// ReviewAuthorDTO — автор отзыва.
//easyjson:json
type ReviewAuthorDTO struct {
	ID       uint64  `json:"id"`
	Nickname string  `json:"nickname"`
	Avatar   *string `json:"avatar,omitempty"`
}

// ReviewWithAuthorResponse — отзыв с автором.
//easyjson:json
type ReviewWithAuthorResponse struct {
	ID        uint64          `json:"id"`
	Title     *string         `json:"title,omitempty"`
	Rating    int16           `json:"rating"`
	Comment   string          `json:"content"`
	CreatedAt time.Time       `json:"createdAt"`
	Author    ReviewAuthorDTO `json:"author"`
}

//easyjson:json
type ReviewWithAuthorList []ReviewWithAuthorResponse

// BoolResponse используется для ответов типа {"in_trip": true}.
//easyjson:json
type BoolResponse struct {
	InTrip bool `json:"in_trip"`
}

// StringResponse используется для простых строковых ответов {"url": "..."}.
//easyjson:json
type StringResponse struct {
	URL string `json:"url,omitempty"`
}
