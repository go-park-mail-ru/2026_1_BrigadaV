package models

import "time"

type Trip struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedBy   uint64     `json:"created_by"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TripMember struct {
	TripID   uint64    `json:"trip_id"`
	UserID   uint64    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
