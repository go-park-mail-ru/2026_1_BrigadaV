package models

import "time"

type Category struct {
	ID          uint64
	Name        string
	Description string
	CreatedAt   time.Time
}
