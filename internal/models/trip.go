package models

import "time"

type Trip struct {
	ID          uint64
	Title       string
	Description string
	StartDate   *time.Time
	EndDate     *time.Time
	CreatedBy   uint64
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
