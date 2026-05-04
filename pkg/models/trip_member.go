package models

import "time"

type TripMember struct {
	TripID   uint64
	UserID   uint64
	Role     string
	JoinedAt time.Time
}
