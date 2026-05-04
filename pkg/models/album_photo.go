package models

import "time"

type AlbumPhoto struct {
	AlbumID    uint64
	PhotoID    uint64
	OrderIndex int16
	CreatedAt  time.Time
	FilePath   string // url/path из таблицы photo, заполняется при GetPhotos
}
