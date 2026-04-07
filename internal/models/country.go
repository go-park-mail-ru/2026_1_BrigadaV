package models

import "time"

type Country struct {
	ID        uint64
	Name      string
	CreatedAt time.Time
}
