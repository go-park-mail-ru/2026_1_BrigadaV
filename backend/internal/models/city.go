package models

import "time"

type City struct {
	ID        int64
	Name      string
	CountryID int64
	CreatedAt time.Time
}
