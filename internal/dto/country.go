//go:generate easyjson -all country.go

package dto

import "time"

//easyjson:json
type CountryResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

//easyjson:json
type CountryResponseList []CountryResponse

//easyjson:json
type LocalityResponse struct {
	ID        uint64   `json:"id"`
	Name      string   `json:"name"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

//easyjson:json
type CountryWithLocalitiesResponse struct {
	ID         uint64             `json:"id"`
	Name       string             `json:"name"`
	Localities []LocalityResponse `json:"localities"`
}
