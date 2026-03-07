package models

import "time"

type Session struct {
	ID           int64
	UserID       int64
	SessionToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}
