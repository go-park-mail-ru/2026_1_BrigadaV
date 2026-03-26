package models

import "time"

type Review struct {
	ID        uint64     `json:"id"`
	UserID    uint64     `json:"user_id"`
	PlaceID   uint64     `json:"place_id"`
	Rating    int16      `json:"rating"`
	Comment   string     `json:"comment,omitempty"`
	VisitDate *time.Time `json:"visit_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
