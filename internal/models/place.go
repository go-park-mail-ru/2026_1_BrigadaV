package models

import "time"

type Place struct {
	ID          uint64
	Name        string
	Description string
	Price       int64
	Locality    Locality
	Category    Category
	Photos      []PlacePhoto
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PlacePhoto struct {
	ID       uint64
	PlaceID  uint64
	FilePath string
	IsMain   bool
}
