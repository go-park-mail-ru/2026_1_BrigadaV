package models

import "time"

type Photo struct {
	ID        uint64
	FilePath  string
	CreatedAt time.Time
}
