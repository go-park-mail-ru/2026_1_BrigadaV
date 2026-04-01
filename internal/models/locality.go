package models

type Locality struct {
	ID        uint64
	Name      string
	Country   string
	Latitude  *float64
	Longitude *float64
}
