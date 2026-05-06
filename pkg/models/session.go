package models

import "time"

type Session struct {
	ID        uint64
	UserID    uint64
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}
