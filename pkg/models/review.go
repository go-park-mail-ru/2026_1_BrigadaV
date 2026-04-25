package models

import "time"

type Review struct {
	ID        uint64
	UserID    uint64
	PlaceID   uint64
	Title     *string
	Rating    int16
	Comment   string
	VisitDate *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ReviewWithAuthor struct {
	ID        uint64    `json:"id"`
	Title     *string   `json:"title,omitempty"`
	Rating    int16     `json:"rating"`
	Comment   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	Author    struct {
		ID       uint64  `json:"id"`
		Nickname string  `json:"nickname"`
		Avatar   *string `json:"avatar,omitempty"`
	} `json:"author"`
}
