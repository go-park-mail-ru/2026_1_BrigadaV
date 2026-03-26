package models

type Locality struct {
	ID        uint64   `json:"id"`
	Name      string   `json:"name"`
	Country   string   `json:"country"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}
