//go:generate easyjson -all country.go

package dto

import "time"

//easyjson:json
type CountryResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
