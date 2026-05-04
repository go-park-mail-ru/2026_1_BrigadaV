package models

import "time"

type AlbumPhoto struct {
	AlbumID    uint64
	PhotoID    uint64
	OrderIndex int16
	CreatedAt  time.Time
}
