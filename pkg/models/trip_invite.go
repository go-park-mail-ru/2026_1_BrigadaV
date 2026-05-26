package models

import "time"

type TripInvite struct {
	ID        uint64
	TripID    uint64
	Token     string
	Role      string // "editor" или "viewer"
	IsOneTime bool
	ExpiresAt *time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
	CreatedBy uint64
}
